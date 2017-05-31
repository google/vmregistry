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
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	libvirt "github.com/libvirt/libvirt-go"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type libvirtMac struct {
	Address string `xml:"address,attr"`
}

type libvirtInterface struct {
	Mac libvirtMac `xml:"mac"`
}

type libvirtDevice struct {
	Interface []libvirtInterface `xml:"interface"`
}

type libvirtMetadata struct {
	MLP mlpMetadata `xml:"mlp"`
}

type libvirtDomain struct {
	Devices  libvirtDevice   `xml:"devices"`
	Metadata libvirtMetadata `xml:"metadata"`
}

type mlpMetadata struct {
	IP string `xml:"ip"`
}

func traceListAllDomains(ctx context.Context, conn *libvirt.Connect) ([]libvirt.Domain, error) {
	sp, _ := opentracing.StartSpanFromContext(ctx, "libvirt.ListAllDomains")
	sp.SetTag("component", "libvirt")
	sp.SetTag("span.kind", "client")
	defer sp.Finish()

	domains, err := conn.ListAllDomains(0)

	if err != nil {
		sp.SetTag("error", true)
		return nil, grpc.Errorf(codes.Unavailable, "failed to get domains: %v", err)
	}

	sp.LogFields(log.Int("domains", len(domains)))
	return domains, nil
}

func traceDomainGetName(ctx context.Context, dom libvirt.Domain) (string, error) {
	sp, _ := opentracing.StartSpanFromContext(ctx, "libvirt.domain.GetName")
	sp.SetTag("component", "libvirt")
	sp.SetTag("span.kind", "client")
	defer sp.Finish()

	name, err := dom.GetName()

	if err != nil {
		sp.SetTag("error", true)
		return "", grpc.Errorf(codes.Unavailable, "failed to get domain name: %v", err)
	}
	return name, nil
}

func traceDomainGetXMLDesc(ctx context.Context, dom libvirt.Domain) (string, error) {
	sp, _ := opentracing.StartSpanFromContext(ctx, "libvirt.domain.GetXMLDesc")
	sp.SetTag("component", "libvirt")
	sp.SetTag("span.kind", "client")
	defer sp.Finish()

	xml, err := dom.GetXMLDesc(libvirt.DOMAIN_XML_INACTIVE)

	if err != nil {
		sp.SetTag("error", true)
		return "", grpc.Errorf(codes.Unavailable, "failed to get domain xml: %v", err)
	}
	return xml, nil
}
