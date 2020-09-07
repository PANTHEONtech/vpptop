/*
 * Copyright (c) 2020 Cisco and/or its affiliates.
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

package api

import (
	govppapi "git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
	"git.fd.io/govpp.git/proxy"
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
)

// VppClient implements VPP-Agent client interface
type VppClient struct {
	vppConn       *core.Connection
	statsConn     govppapi.StatsProvider
	client        *proxy.Client
	vppInfo       VPPInfo
	apiChan       govppapi.Channel
	binapiVersion vpp.Version
}

// NewVppClient returns VPP client connected to the VPP via the shared memory
func NewVppClient(vppConn *core.Connection, statsConn govppapi.StatsProvider) *VppClient {
	return &VppClient{
		vppConn:   vppConn,
		statsConn: statsConn,
	}
}

// NewProxyClient returns VPP client which is connected to the VPP dialing remote proxy server
func NewProxyClient(client *proxy.Client, statsConn govppapi.StatsProvider) *VppClient {
	return &VppClient{
		client:    client,
		statsConn: statsConn,
	}
}

func (c VppClient) NewAPIChannel() (govppapi.Channel, error) {
	if c.client != nil {
		return c.client.NewBinapiClient()
	}
	return c.vppConn.NewAPIChannel()
}

func (c *VppClient) CheckCompatiblity(msgs ...govppapi.Message) error {
	if c.apiChan == nil {
		ch, err := c.NewAPIChannel()
		if err != nil {
			return err
		}
		c.apiChan = ch
	}
	return c.apiChan.CheckCompatiblity(msgs...)
}

func (c *VppClient) Stats() govppapi.StatsProvider {
	return c.statsConn
}

func (c *VppClient) IsPluginLoaded(plugin string) bool {
	for _, p := range c.vppInfo.Plugins {
		if p.Name == plugin {
			return true
		}
	}
	return false
}

func (c *VppClient) BinapiVersion() vpp.Version {
	return vpp.Version(c.vppInfo.Version)
}

func (c *VppClient) OnReconnect(_ func()) {
	// no-op
}

func (c *VppClient) Connection() govppapi.Connection {
	return c.vppConn
}

// SetInfo about the connected VPP
func (c *VppClient) SetInfo(vppInfo VPPInfo) {
	c.vppInfo = vppInfo
}

// Disconnect from the VPP
func (c *VppClient) Disconnect() {
	if c.vppConn != nil {
		c.vppConn.Disconnect()
	}
}

func (c *VppClient) Close() {
	if c.apiChan != nil {
		c.apiChan.Close()
	}
}
