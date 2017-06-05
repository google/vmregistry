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
	"flag"
	"fmt"

	"golang.org/x/net/context"

	"github.com/google/credstore/client"
	microClient "github.com/google/go-microservice-helpers/client"
	pb "github.com/google/lvmd/proto"
)

var (
	lvmMirrors = flag.Int("lvm-mirrors", 0, "lvm mirrors count, 0 to disable")
)

type LVMStorage struct {
	client   pb.LVMClient
	vg       string
	lvmToken string
}

func newLVMClient(lvmAddress string, lvmCA string) (pb.LVMClient, error) {
	conn, err := microClient.NewGRPCConn(lvmAddress, lvmCA, "", "")
	if err != nil {
		return nil, err
	}
	cli := pb.NewLVMClient(conn)
	return cli, nil
}

func NewLVMStorage(lvmAddress string, lvmCA string, vg string, lvmToken string) (StorageManager, error) {
	client, err := newLVMClient(lvmAddress, lvmCA)
	if err != nil {
		return nil, err
	}

	return LVMStorage{client, vg, lvmToken}, nil
}

func (s LVMStorage) authContext(ctx context.Context) context.Context {
	if s.lvmToken == "" {
		return ctx
	}

	return client.WithBearerToken(ctx, s.lvmToken)
}

func (s LVMStorage) CreateStorage(ctx context.Context, name string, size uint64, sourceImage string) error {
	ctx = s.authContext(ctx)

	_, err := s.client.CreateLV(ctx, &pb.CreateLVRequest{
		VolumeGroup: s.vg,
		Name:        name,
		Size:        size,
		Mirrors:     uint32(*lvmMirrors),
		Tags:        []string{"vm"},
	})
	if err != nil {
		return err
	}

	s.client.CloneLV(ctx, &pb.CloneLVRequest{
		SourceName: sourceImage,
		DestName:   s.StorageBlockDevice(name),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s LVMStorage) RemoveStorage(ctx context.Context, name string) error {
	ctx = s.authContext(ctx)

	_, err := s.client.RemoveLV(ctx, &pb.RemoveLVRequest{
		VolumeGroup: s.vg,
		Name:        name,
	})

	return err
}

func (s LVMStorage) StorageBlockDevice(name string) string {
	return fmt.Sprintf("/dev/%s/%s", s.vg, name)
}
