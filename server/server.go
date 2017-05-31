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
	"encoding/xml"

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
	return dom.Metadata.MLP.IP
}

// Server is GRPC server.
type Server struct {
	conn *libvirt.Connect
}

// NewServer creates a new server instance.
func NewServer(conn *libvirt.Connect) Server {
	return Server{conn: conn}
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
