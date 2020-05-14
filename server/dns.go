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
	"strings"

	"github.com/mrturkmen06/vmregistry/powerdns"
)

type DnsClient struct {
	cli    *powerdns.PowerDNS
	domain string
}

func NewDNSClient(baseURL string, domain string, apikey string) *DnsClient {
	if !strings.HasSuffix(domain, ".") {
		domain = domain + "."
	}
	return &DnsClient{cli: powerdns.New(baseURL, "localhost", domain, apikey), domain: domain}
}

func (c DnsClient) Add(name string, address string) error {
	_, err := c.cli.AddRecord(name+"."+c.domain, "A", 300, []string{address})
	return err
}

func (c DnsClient) Remove(name string, address string) error {
	_, err := c.cli.DeleteRecord(name+"."+c.domain, "A", 300, []string{address})
	return err
}
