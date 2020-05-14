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
	"context"
	"os"

	"github.com/golang/glog"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	pb "github.com/mrturkmen06/vmregistry/api"
)

var (
	createVMName        string
	createVMMem         uint64
	createVMCores       uint32
	createVMSize        uint64
	createVMSourceImage string
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		createVMSize = createVMSize * 1024 * 1024 * 1024

		initCredStoreSession()

		ctx, err := vmregistryContext(context.Background())
		if err != nil {
			glog.Fatalf("failed to acquire a client vmregistry context: %v", err)
		}

		client, err := newClient()
		if err != nil {
			glog.Fatalf("failed to create a client: %v", err)
		}

		vm, err := client.Create(ctx, &pb.CreateRequest{
			Name:        createVMName,
			Mem:         createVMMem,
			Cores:       createVMCores,
			Size:        createVMSize,
			SourceImage: createVMSourceImage,
		})
		if err != nil {
			glog.Fatalf("failed to create VM: %v", err)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "IP"})

		table.Append([]string{vm.Name, vm.Ip})

		table.Render()
	},
}

func init() {
	RootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	createCmd.Flags().StringVar(&createVMName, "name", "", "vm name")
	createCmd.Flags().Uint64Var(&createVMMem, "mem", 1, "vm memory in GB")
	createCmd.Flags().Uint32Var(&createVMCores, "cores", 1, "vm cores")
	createCmd.Flags().Uint64Var(&createVMSize, "size", 3, "vm disk in GB")
	createCmd.Flags().StringVar(&createVMSourceImage, "source-image", "", "vm source image")
}
