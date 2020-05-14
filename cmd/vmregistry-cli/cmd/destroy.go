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

	"github.com/golang/glog"
	"github.com/spf13/cobra"

	pb "github.com/mrturkmen06/vmregistry/api"
)

// destroyCmd represents the create command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			glog.Fatalf("destroy needs a name")
		}

		name := args[0]

		initCredStoreSession()

		ctx, err := vmregistryContext(context.Background())
		if err != nil {
			glog.Fatalf("failed to acquire a client vmregistry context: %v", err)
		}

		client, err := newClient()
		if err != nil {
			glog.Fatalf("failed to create a client: %v", err)
		}

		_, err = client.Destroy(ctx, &pb.DestroyRequest{
			Name: name,
		})
		if err != nil {
			glog.Fatalf("failed to destroy VM: %v", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(destroyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// destroyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// destroyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
