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

package main

import (
	"context"
	"flag"
	"html/template"
	"io/ioutil"
	"net"

	pb "github.com/google/vmregistry/api"
	"github.com/google/vmregistry/server"
	"github.com/google/vmregistry/web"

	"github.com/golang/glog"
	"github.com/google/go-microservice-helpers/server"
	"github.com/google/go-microservice-helpers/tracing"
	"github.com/libvirt/libvirt-go"
)

var (
	libvirtURI = flag.String("libvirt-uri", "", "libvirt connection uri")
	vmTemplate = flag.String("vm-template-file", "", "path to libvirt xml template file to be used for vm creation")
	vmNet      = flag.String("vm-net", "", "A subnet for VM ip address generation")
	vmVG       = flag.String("vm-vg", "", "lvm volume group for storage")

	lvmdAddress = flag.String("lvmd-address", "", "lvmd grpc address")
	lvmdCA      = flag.String("lvmd-ca", "", "lvmd server ca")

	dnsAPIURL = flag.String("pdns-api-url", "", "PowerDNS base URL")
	dnsZone   = flag.String("pdns-zone", "", "Zone to host VMs")
	dnsAPIKey = flag.String("pdns-api-key", "", "PowerDNS API Key")
)

func main() {
	flag.Parse()
	defer glog.Flush()

	conn, err := libvirt.NewConnect(*libvirtURI)
	if err != nil {
		glog.Fatalf("failed to connect to libvirt: %v", err)
	}

	err = tracing.InitTracer(*serverhelpers.ListenAddress, "vmregistry")
	if err != nil {
		glog.Fatalf("failed to init tracing interface: %v", err)
	}

	_, net, err := net.ParseCIDR(*vmNet)
	if err != nil {
		glog.Fatalf("failed to parse vm net: %v", err)
	}

	grpcServer, credstoreClient, err := serverhelpers.NewServer()
	if err != nil {
		glog.Fatalf("failed to init GRPC server: %v", err)
	}
	if credstoreClient == nil {
		glog.Fatalf("failed to init credstore")
	}

	lvmSessionTok, err := credstoreClient.GetTokenForRemote(context.Background(), *lvmdAddress)
	if err != nil {
		glog.Fatalf("failed to get lvmd token: %v", err)
	}

	storage, err := server.NewLVMStorage(*lvmdAddress, *lvmdCA, *vmVG, lvmSessionTok)
	if err != nil {
		glog.Fatalf("failed to create connection to lvmd: %v", err)
	}

	tpl, err := ioutil.ReadFile(*vmTemplate)
	if err != nil {
		glog.Fatalf("failed to load vm template: %v", err)
	}
	var xmlTemplate = template.Must(template.New("domain").Parse(string(tpl)))

	dnsCli := server.NewDNSClient(*dnsAPIURL, *dnsZone, *dnsAPIKey)

	svr := server.NewServer(conn, storage, net, dnsCli, xmlTemplate)

	pb.RegisterVMRegistryServer(grpcServer, &svr)

	statusHandler := web.NewStatusHandler(&svr)

	err = serverhelpers.ListenAndServe(grpcServer, statusHandler)
	if err != nil {
		glog.Fatalf("failed to serve: %v", err)
	}
}
