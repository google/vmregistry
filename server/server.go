/*

Copyright 2017 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

package server

import (
	"bytes"
	"encoding/xml"
	"html/template"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "github.com/google/vmregistry/api"

	"github.com/golang/glog"
	libvirt "github.com/libvirt/libvirt-go"
)

func extractMACs(dom libvirtDomain) []string {
	macs := []string{}
	for _, i := range dom.Devices.Interface {
		macs = append(macs, i.Mac.Address)
	}

	return macs
}

func extractIP(dom libvirtDomain) string {
	return dom.Metadata.VMRegistry.IP
}

type StorageManager interface {
	CreateStorage(ctx context.Context, name string, size uint64, sourceImage string) error
	RemoveStorage(ctx context.Context, name string) error
	StorageBlockDevice(name string) string
}

// Server is GRPC server.
type Server struct {
	conn    *libvirt.Connect
	storage StorageManager
	vmNet   *net.IPNet
	dnsCli  *DnsClient

	xmlTemplate *template.Template
}

// NewServer creates a new server instance.
func NewServer(conn *libvirt.Connect, storage StorageManager, vmNet *net.IPNet, dnsCli *DnsClient, xmlTemplate *template.Template) Server {
	return Server{
		conn:        conn,
		storage:     storage,
		vmNet:       vmNet,
		dnsCli:      dnsCli,
		xmlTemplate: xmlTemplate,
	}
}

// List is GRPC handler for List API.
func (s Server) List(ctx context.Context, req *pb.ListVMRequest) (*pb.ListVMReply, error) {
	domains, err := traceListAllDomains(ctx, s.conn)
	if err != nil {
		return nil, err
	}

	repl := &pb.ListVMReply{}
	repl.Vms = make([]*pb.VM, len(domains))

	for i, d := range domains {
		name, err := traceDomainGetName(ctx, d)
		if err != nil {
			return nil, err
		}

		domXML, err := traceDomainGetXMLDesc(ctx, d)
		if err != nil {
			return nil, err
		}

		// TODO(farcaller): fails to load this
		// metadataXML, err := d.GetMetadata(libvirt.DOMAIN_METADATA_ELEMENT, MLPNamespace, libvirt.DOMAIN_AFFECT_LIVE)

		dom := libvirtDomain{}
		err = xml.Unmarshal([]byte(domXML), &dom)
		if err != nil {
			return nil, grpc.Errorf(codes.Internal, "failed to parse domain xml: %v", err)
		}

		macs := extractMACs(dom)
		if len(macs) != 1 {
			glog.Warningf("strange mac count on %s: %v", name, macs)
		}

		mac := macs[0]
		ip := extractIP(dom)
		if ip == "" {
			glog.Warningf("failed to get ip for node %s", name)
		}

		repl.Vms[i] = &pb.VM{
			Name: name,
			Ip:   ip,
			Mac:  mac,
		}
	}

	return repl, nil
}

// Find is GRPC handler for Find API.
func (s Server) Find(ctx context.Context, req *pb.FindRequest) (*pb.VM, error) {
	if req.FindBy == pb.FindRequest_UNSPECIFIED {
		return nil, grpc.Errorf(codes.InvalidArgument, "search criteria not specified")
	}

	domains, err := traceListAllDomains(ctx, s.conn)
	if err != nil {
		return nil, err
	}

	for _, d := range domains {
		name, err := traceDomainGetName(ctx, d)
		if err != nil {
			continue
		}

		domXML, err := traceDomainGetXMLDesc(ctx, d)
		if err != nil {
			continue
		}

		dom := libvirtDomain{}
		err = xml.Unmarshal([]byte(domXML), &dom)
		if err != nil {
			continue
		}

		macs := extractMACs(dom)
		if len(macs) != 1 {
			glog.Warningf("strange mac count on %s: %v", name, macs)
		}

		mac := macs[0]
		ip := extractIP(dom)

		repl := &pb.VM{
			Name: name,
			Ip:   ip,
			Mac:  mac,
		}

		if req.FindBy == pb.FindRequest_IP && ip == req.Value {
			return repl, nil
		}

		if req.FindBy == pb.FindRequest_MAC && mac == req.Value {
			return repl, nil
		}
	}

	return nil, grpc.Errorf(codes.NotFound, "ip not found")
}

// Create is GRPC handler for Create API.
func (s Server) Create(ctx context.Context, in *pb.CreateRequest) (*pb.VM, error) {
	name := in.GetName()
	if name == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "name not specified")
	}
	mem := in.GetMem()
	if mem == 0 {
		return nil, grpc.Errorf(codes.InvalidArgument, "mem not specified")
	}
	cores := in.GetCores()
	if cores == 0 {
		return nil, grpc.Errorf(codes.InvalidArgument, "cores not specified")
	}
	size := in.GetSize()
	if size == 0 {
		return nil, grpc.Errorf(codes.InvalidArgument, "size not specified")
	}
	sourceImage := in.GetSourceImage()
	if sourceImage == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "sourceImage not specified")
	}

	err := s.storage.CreateStorage(ctx, name, size, sourceImage)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to create storage: %v", err)
	}

	var ip net.IP
	for i := 0; i < 10; i++ {
		tryip := generateIPv4(s.vmNet)
		searchReq := &pb.FindRequest{
			FindBy: pb.FindRequest_IP,
			Value:  tryip.String(),
		}
		_, err := s.Find(ctx, searchReq)
		code := grpc.Code(err)
		if code == codes.NotFound {
			ip = tryip
			break
		}
	}
	if len(ip) == 0 {
		return nil, grpc.Errorf(codes.Unavailable, "failed to generate a new ip after 10 attempts")
	}

	var domBuffer bytes.Buffer
	s.xmlTemplate.Execute(&domBuffer, struct {
		Name     string
		Memory   uint64
		Cores    uint32
		DiskPath string
		IP       string
	}{
		Name:     name,
		Memory:   in.GetMem(),
		Cores:    in.GetCores(),
		DiskPath: s.storage.StorageBlockDevice(name),
		IP:       ip.String(),
	})
	domXML := domBuffer.String()

	d, err := s.conn.DomainDefineXML(domXML)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to define vm: %v", err)
	}

	err = d.Create()
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to create vm: %v", err)
	}

	err = s.dnsCli.Add(name, ip.String())
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to update dns record: %v", err)
	}

	return &pb.VM{
		Name: name,
		Ip:   ip.String(),
		Mac:  "FIXME",
	}, nil
}

// Destroy is GRPC handler for Destroy API.
func (s Server) Destroy(ctx context.Context, in *pb.DestroyRequest) (*pb.DestroyReply, error) {
	name := in.GetName()
	if name == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "name not specified")
	}

	dom, err := traceGetDomainByName(ctx, s.conn, name)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to lookup vm: %v", err)
	}

	domXML, err := traceDomainGetXMLDesc(ctx, *dom)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to get vm xml: %v", err)
	}

	domData := libvirtDomain{}
	err = xml.Unmarshal([]byte(domXML), &domData)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to parse domain xml: %v", err)
	}

	ip := extractIP(domData)
	if ip == "" {
		return nil, grpc.Errorf(codes.Internal, "failed to get ip for node %s", name)
	}

	err = s.dnsCli.Remove(name, ip)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to update dns record: %v", err)
	}

	err = dom.Destroy()
	if err != nil {
		glog.Infof("failed to destroy vm: %v, continuing with undefining", err)
	}

	err = dom.Undefine()
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to undefine vm: %v", err)
	}

	err = s.storage.RemoveStorage(ctx, name)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "failed to remove vm storage: %v", err)
	}

	return &pb.DestroyReply{}, nil
}
