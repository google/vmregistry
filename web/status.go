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

package web

import (
	"html/template"
	"net/http"

	pb "github.com/mrturkmen06/vmregistry/api"
	"github.com/mrturkmen06/vmregistry/server"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
)

var promHandler = prometheus.Handler()

// StatusHandler is default http handler
type StatusHandler struct {
	svr *server.Server
}

// NewStatusHandler creates a StatusHandler.
func NewStatusHandler(svr *server.Server) StatusHandler {
	return StatusHandler{svr: svr}
}

func (h StatusHandler) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	if rq.URL.Path == "/metrics" {
		promHandler.ServeHTTP(w, rq)
	}

	if rq.URL.Path != "/" {
		return
	}

	repl, err := h.svr.List(rq.Context(), &pb.ListVMRequest{})
	if err != nil {
		glog.Errorf("falied to list vms: %v", err)
		return
	}

	err = pageTemplate.Execute(w, repl)
	if err != nil {
		glog.Errorf("falied to render status page: %v", err)
		return
	}
}

const pageTemplateText = `
<html>
<body>
	<h1>vmregistry</h1>
	<table border=1>
		<tr>
			<th>Name</th>
			<th>IP</th>
			<th>MAC</th>
		</tr>
		{{range .Vms}}
		<tr>
			<td>{{.Name}}</td>
			<td>{{.Ip}}</td>
			<td>{{.Mac}}</td>
		</tr>
		{{end}}
	</table>
</body>
</html>
`

var pageTemplate = template.Must(template.New("page").Parse(pageTemplateText))
