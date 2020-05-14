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

package cmd

import (
	"strings"

	"github.com/golang/glog"
	"github.com/google/credstore/client"
	microClient "github.com/google/go-microservice-helpers/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/mrturkmen06/vmregistry/api"
)

var (
	credStoreSessionTok string
	credStoreConn       *grpc.ClientConn
)

func newClient() (pb.VMRegistryClient, error) {
	conn, err := microClient.NewGRPCConn(vmregistryGrpcURL, vmregistryGrpcCA, "", "")
	if err != nil {
		return nil, err
	}
	cli := pb.NewVMRegistryClient(conn)
	return cli, nil
}

func initCredStoreSession() {
	if credStoreAddress != "" {
		appTok, err := client.GetAppToken()
		if err != nil {
			glog.Fatalf("failed to get app token: %v", err)
		}
		credStoreConn, err = microClient.NewGRPCConn(credStoreAddress, credStoreCA, "", "")
		if err != nil {
			glog.Fatalf("failed to create connection to credstore: %v", err)
		}
		credStoreSessionTok, err = client.GetAuthToken(context.Background(), credStoreConn, appTok)
		if err != nil {
			glog.Fatalf("failed to get session token: %v", err)
		}
	}
}

func vmregistryContext(ctx context.Context) (context.Context, error) {
	vmregistryGrpcDNSName := strings.Split(vmregistryGrpcURL, ":")[0]
	tok, err := client.GetTokenForRemote(context.Background(), credStoreConn, credStoreSessionTok, vmregistryGrpcDNSName)
	if err != nil {
		return nil, err
	}

	return client.WithBearerToken(ctx, tok), nil
}
