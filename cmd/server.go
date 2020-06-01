/*
Copyright © 2020 NAME HERE christian@slashdevops.com

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
	"fmt"

	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var (
	confFile  string
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "A brief description of your command",
		Long:  `A longer description that spans `,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("server called")
		},
	}
)

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().StringVar(&confFile, "config", "", "config file (default is $HOME/server.yaml)")
	serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
