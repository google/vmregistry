/*

Copyright 2017 Google Inc.

Heavily based on https://github.com/waynz0r/powerdns.

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

package powerdns

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/dghubble/sling"
)

// Error strct
type Error struct {
	Message string `json:"error"`
}

// Error Returns
func (e Error) Error() string {
	return fmt.Sprintf("%v", e.Message)
}

// CombinedRecord strct
type CombinedRecord struct {
	Name    string
	Type    string
	TTL     int
	Records []string
}

// Zone struct
type Zone struct {
	ID             string `json:"id"`
	URL            string `json:"url"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	DNSsec         bool   `json:"dnssec"`
	Serial         int    `json:"serial"`
	NotifiedSerial int    `json:"notified_serial"`
	LastCheck      int    `json:"last_check"`
	Records        []struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		TTL      int    `json:"ttl"`
		Priority int    `json:"priority"`
		Disabled bool   `json:"disabled"`
		Content  string `json:"content"`
	} `json:"records"`
}

// Record struct
type Record struct {
	Disabled bool   `json:"disabled"`
	Content  string `json:"content"`
}

// RRset struct
type RRset struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	TTL        int      `json:"ttl"`
	ChangeType string   `json:"changetype"`
	Records    []Record `json:"records"`
}

// RRsets struct
type RRsets struct {
	Sets []RRset `json:"rrsets"`
}

// PowerDNS struct
type PowerDNS struct {
	scheme   string
	hostname string
	port     string
	vhost    string
	domain   string
	apikey   string
}

// New returns a new PowerDNS
func New(baseURL string, vhost string, domain string, apikey string) *PowerDNS {
	if vhost == "" {
		vhost = "localhost"
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		log.Fatalf("%s is not a valid url: %v", baseURL, err)
	}
	hp := strings.Split(u.Host, ":")
	hostname := hp[0]
	var port string
	if len(hp) > 1 {
		port = hp[1]
	} else {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	return &PowerDNS{
		scheme:   u.Scheme,
		hostname: hostname,
		port:     port,
		vhost:    vhost,
		domain:   domain,
		apikey:   apikey,
	}
}

// AddRecord ...
func (p *PowerDNS) AddRecord(name string, recordType string, ttl int, content []string) (*Zone, error) {

	zone, err := p.ChangeRecord(name, recordType, ttl, content, "UPSERT")

	return zone, err
}

// DeleteRecord ...
func (p *PowerDNS) DeleteRecord(name string, recordType string, ttl int, content []string) (*Zone, error) {

	zone, err := p.ChangeRecord(name, recordType, ttl, content, "DELETE")

	return zone, err
}

// ChangeRecord ...
func (p *PowerDNS) ChangeRecord(name string, recordType string, ttl int, content []string, action string) (*Zone, error) {

	Record := new(CombinedRecord)
	Record.Name = name
	Record.Type = recordType
	Record.TTL = ttl
	Record.Records = content

	zone, err := p.patchRRset(*Record, action)

	return zone, err
}

func (p *PowerDNS) patchRRset(record CombinedRecord, action string) (*Zone, error) {

	Set := RRset{Name: record.Name, Type: record.Type, ChangeType: "REPLACE", TTL: record.TTL}

	if action == "DELETE" {
		Set.ChangeType = "DELETE"
	}

	var R Record

	for _, rec := range record.Records {
		R = Record{Content: rec}
		Set.Records = append(Set.Records, R)
	}

	json := RRsets{}
	json.Sets = append(json.Sets, Set)

	error := new(Error)
	zone := new(Zone)

	resp, err := p.getSling().Patch(p.domain).BodyJSON(json).Receive(zone, error)

	if err == nil && resp.StatusCode >= 400 {
		error.Message = strings.Join([]string{resp.Status, error.Message}, " ")
		return nil, error
	}

	return zone, err
}

func (p *PowerDNS) getSling() *sling.Sling {

	u := new(url.URL)
	u.Host = p.hostname + ":" + p.port
	u.Scheme = p.scheme
	u.Path = "/api/v1/servers/" + p.vhost + "/zones/"

	Sling := sling.New().Base(u.String())

	Sling.Set("X-API-Key", p.apikey)

	return Sling
}
