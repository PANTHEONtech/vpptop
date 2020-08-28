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

package command

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"git.fd.io/govpp.git/adapter/socketclient"
	"git.fd.io/govpp.git/adapter/statsclient"
	"git.fd.io/govpp.git/proxy"
	"github.com/ligato/cn-infra/logging"
	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node <nodeName>",
	Short: "Collects vpp statistics from the specified node",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("no node specified")
		}

		logs, err := os.Create("remote.log")
		if err != nil {
			return fmt.Errorf("error occured while creating file: %v", err)
		}

		defer logs.Close()

		logging.DefaultLogger.SetOutput(logs)

		kubeconfig, err := cmd.Flags().GetString("kubeconfig")
		if err != nil {
			return err
		}

		ipaddr, found := resolveNode(kubeconfig, args[0])
		if found {
			return startClient("", ipaddr+":"+"7878", logs)
		}

		log.Println("failed to resolve addr:", args[0])

		rAddr, err := cmd.Flags().GetString("addr")
		if err != nil {
			return err
		}

		log.Println("trying to connect to a local server at:", rAddr)

		for i := 0; i < 3; i++ {
			if _, err = proxy.Connect(rAddr); err == nil {
				break
			}
			time.Sleep(1 * time.Second)
		}

		if err != nil {
			log.Println("no server found")
			log.Println("starting local server at:", rAddr)

			binapiSocket, err := cmd.Flags().GetString("binapi-socket")
			if err != nil {
				return err
			}

			statsSocket, err := cmd.Flags().GetString("stats-socket")
			if err != nil {
				return err
			}

			go func() {
				p, err := proxy.NewServer()
				if err != nil {
					log.Fatalln("creating local server failed")
				}

				statsAdapter := statsclient.NewStatsClient(statsSocket)
				binapiAdapter := socketclient.NewVppClient(binapiSocket)

				if err := p.ConnectStats(statsAdapter); err != nil {
					log.Fatalln("connecting to stats failed:", err)
				}

				defer p.DisconnectStats()

				if err := p.ConnectBinapi(binapiAdapter); err != nil {
					log.Fatalln("connecting to binapi failed:", err)
				}

				defer p.DisconnectBinapi()

				p.ListenAndServe(rAddr)
			}()
		}

		return startClient("", rAddr, logs)
	},
}

func init() {
	if home := homeDir(); home != "" {
		nodeCmd.Flags().StringP("kubeconfig", "c", filepath.Join(home, ".kube", "config"), "(optional) absolute path to kubeconfig")
	} else {
		nodeCmd.Flags().StringP("kubeconfig", "c", "", "absolute path to the kubeconfig")
	}

	nodeCmd.Flags().String("binapi-socket", socketclient.DefaultSocketName, "Path to VPP binapi socket")
	nodeCmd.Flags().String("stats-socket", statsclient.DefaultSocketName, "Path to VPP stats socket")
	nodeCmd.Flags().String("addr", ":9191", "Address on which proxy serves RPC.")
	rootCmd.AddCommand(nodeCmd)
}
