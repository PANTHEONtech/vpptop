/*
 * Copyright (c) 2019 PANTHEON.tech.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"log"

	"github.com/PantheonTechnologies/vpptop/stats"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vpptop",
	Short: "vpptop is a dynamic terminal user interface for VPP metrics",
	Long: `vpptop is a go implementation of real-time viewer for VPP metrics
shown in dynamic terminal user interface. vpptop supports the following metrics:

Interface:      stats: RX/TX packets/bytes, packet errors/drops/punts/IPv4...
Node stats:     clocks, vectors, calls, suspends...
Error counters: node, reason...
Memory usage:   free, used...
Thread info:    name, type, PID...`,

	RunE: func(cmd *cobra.Command, args []string) error {
		socket, err := cmd.Flags().GetString("socket")
		if err != nil {
			return err
		}
		logFile, err := cmd.Flags().GetString("log")
		if err != nil {
			return err
		}
		return startClient(socket, "", logFile)
	},
}

func init() {
	rootCmd.Flags().StringP("socket", "s", stats.DefaultSocket, "vpp stats segment socket")
	rootCmd.Flags().StringP("log", "l", "vpptop.log", "Log file")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
