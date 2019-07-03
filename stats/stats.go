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

package stats

import (
	"fmt"
	"net"
	"strings"

	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/adapter"
	"git.fd.io/govpp.git/adapter/statsclient"
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
	"github.com/PantheonTechnologies/vpptop/bin_api/interfaces"
	"github.com/PantheonTechnologies/vpptop/bin_api/ip"
	"github.com/PantheonTechnologies/vpptop/bin_api/vpe"
)

const (
	stateUp   = "up"
	stateDown = "down"
)

var (
	DefaultSocket = adapter.DefaultStatsSocket
)

type VPP struct {
	client    adapter.StatsAPI
	statsConn *core.StatsConnection
	vppConn   *core.Connection
	apiChan   api.Channel

	IfCache []Interface
}

// Node extends the counters from the api.NodeCounters
// to also include Vectors/Calls.
type Node struct {
	api.NodeCounters
	VC float64
}

// Interface extends the counters from the api.InterfaceCounters
// to also include Ipv4 addresses, IPv6 addresses, state and Mtu.
type Interface struct {
	api.InterfaceCounters
	IPv4  []string
	IPv6  []string
	State string
	Mtu   []uint32
}

// Error counters.
type Error struct {
	Value    uint64
	NodeName string
	Reason   string
}

// Connect establishes a connection to the govpp API.
func (s *VPP) Connect(soc string) error {
	s.client = statsclient.NewStatsClient(soc)

	var err error
	s.statsConn, err = core.ConnectStats(s.client)
	if err != nil {
		return fmt.Errorf("connection to stats api failed: %v", err)
	}

	s.vppConn, err = govpp.Connect("")
	if err != nil {
		return fmt.Errorf("connection to govpp failed: %v", err)
	}

	s.apiChan, err = s.vppConn.NewAPIChannel()
	if err != nil {
		return fmt.Errorf("api channel creation failed: %v", err)
	}

	var msgs []api.Message
	msgs = append(msgs, interfaces.AllMessages()...)
	msgs = append(msgs, ip.AllMessages()...)
	msgs = append(msgs, vpe.AllMessages()...)
	err = s.apiChan.CheckCompatiblity(msgs...)
	if err != nil {
		return fmt.Errorf("compatibility check failed: %v", err)
	}

	return nil
}

// Version returns the current vpp version.
func (s *VPP) Version() (string, error) {
	req := &vpe.ShowVersion{}
	reply := &vpe.ShowVersionReply{}

	if err := s.apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	return "VPP version: " + reply.Version + "\n" + reply.BuildDate, nil
}

// Disconnect should be called after Connect, if the connection is no longer needed.
func (s *VPP) Disconnect() {
	s.apiChan.Close()
	s.vppConn.Disconnect()
	s.statsConn.Disconnect()
}

// interfacesDetails returns the details of each interface from the bin_api.
func (s *VPP) interfacesDetails() (map[uint32]*interfaces.SwInterfaceDetails, error) {
	result := make(map[uint32]*interfaces.SwInterfaceDetails)

	req := s.apiChan.SendMultiRequest(&interfaces.SwInterfaceDump{})

	for {
		iface := &interfaces.SwInterfaceDetails{}
		stop, err := req.ReceiveReply(iface)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		result[iface.SwIfIndex] = iface
	}
	return result, nil
}

// ipv4Addresses return all the IPv4 addresses assigned to the interface.
func (s *VPP) ipv4Addresses(ifIndex uint32) ([]string, error) {
	var ipv4Addresses []string
	req := &ip.IPAddressDump{SwIfIndex: ifIndex, IsIPv6: 0}
	reqCtx := s.apiChan.SendMultiRequest(req)
	for {
		ipDetails := &ip.IPAddressDetails{}
		stop, err := reqCtx.ReceiveReply(ipDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		ipv4 := ipDetails.Prefix.Address.Un.GetIP4()
		ipv4Addresses = append(ipv4Addresses, net.IP(ipv4[:]).String())
	}

	return ipv4Addresses, nil
}

// ipv6Addresses returns all IPv6 addresses assigned to the interface.
func (s *VPP) ipv6Addresses(ifIndex uint32) ([]string, error) {
	var ipv6Addresses []string
	req := &ip.IPAddressDump{SwIfIndex: ifIndex, IsIPv6: 1}
	reqCtx := s.apiChan.SendMultiRequest(req)
	for {
		ipDetails := &ip.IPAddressDetails{}
		stop, err := reqCtx.ReceiveReply(ipDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		ipv6 := ipDetails.Prefix.Address.Un.GetIP6()
		ipv6Addresses = append(ipv6Addresses, net.IP(ipv6[:]).String())
	}
	return ipv6Addresses, nil
}

// GetNodes returns per node statistics.
func (s *VPP) GetNodes() ([]Node, error) {
	var result []Node

	stats, err := s.statsConn.GetNodeStats()
	if err != nil {
		return nil, fmt.Errorf("error occured while retrieving node stats: %v", err)
	}

	result = make([]Node, len(stats.Nodes))
	for i, node := range stats.Nodes {
		vc := 0.0
		if node.Vectors != 0 && node.Calls != 0 {
			vc = float64(node.Vectors) / float64(node.Calls)
		}
		result[i] = Node{
			NodeCounters: node,
			VC:           vc,
		}
	}
	return result, nil
}

// GetInterfaces returns per interface statistics.
func (s *VPP) GetInterfaces() ([]Interface, error) {
	var result []Interface

	stats, err := s.statsConn.GetInterfaceStats()
	if err != nil {
		return nil, fmt.Errorf("error occured while retrieving interface stats: %v", err)
	}

	ifaces, err := s.interfacesDetails()
	if err != nil {
		return nil, fmt.Errorf("error occured while dumping interface details: %v", err)
	}

	// stats-api for interfaces returns counters for deleted interfaces which
	// are no longer present via vppctl. Tt was reported as a bug.
	// You can find it here -> https://jira.fd.io/projects/GOVPP/issues/GOVPP-18?filter=allopenissues

	// To bypass this for now, we check if the bin_api contains the interface
	// from the stats-api.
	result = make([]Interface, 0, len(ifaces))
	for _, iface := range stats.Interfaces {
		ifaceDetails, ok := ifaces[iface.InterfaceIndex]
		if !ok {
			continue
		}
		ifaceState := stateDown
		if ifaceDetails.AdminUpDown > 0 {
			ifaceState = stateUp
		}

		// If error with ip occurs, ignore the ip for the interface.
		ifaceIPv4, err := s.ipv4Addresses(iface.InterfaceIndex)
		if err != nil {
			ifaceIPv4 = make([]string, 0)
			//return nil, err
		}
		ifaceIPv6, err := s.ipv6Addresses(iface.InterfaceIndex)
		if err != nil {
			ifaceIPv6 = make([]string, 0)
			//return nil, err
		}

		result = append(result, Interface{
			InterfaceCounters: iface,
			Mtu:               append([]uint32(nil), ifaceDetails.Mtu...),
			IPv4:              append(make([]string, 0, len(ifaceIPv4)), ifaceIPv4...),
			IPv6:              append(make([]string, 0, len(ifaceIPv6)), ifaceIPv6...),
			State:             ifaceState,
		})
	}
	return result, nil
}

// GetErrors returns per error statistics.
func (s *VPP) GetErrors() ([]Error, error) {
	var result []Error

	stats, err := s.statsConn.GetErrorStats("")
	if err != nil {
		return nil, fmt.Errorf("error occured while retrieving error stats: %v", err)
	}

	result = make([]Error, 0)
	for _, counter := range stats.Errors {
		if counter.Value == 0 {
			continue
		}
		s := strings.Split(counter.CounterName, "/")
		result = append(result, Error{
			Value:    counter.Value,
			NodeName: s[0],
			Reason:   s[1],
		})
	}
	return result, nil
}

// ClearIfaceCounters resets the counters for the interface.
func (s *VPP) ClearIfaceCounters(ifaceIndex uint32) error {
	req := &interfaces.SwInterfaceClearStats{SwIfIndex: ifaceIndex}
	reply := &interfaces.SwInterfaceClearStatsReply{}

	if err := s.apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return fmt.Errorf("request failed: %v", err)
	}

	return nil
}

// ClearRuntimeCounters clears the runtime counters for nodes.
func (s *VPP) ClearRuntimeCounters() error {
	// TODO: find if there exits a bin_api call for clearing node counters 'casue this seems to be slow compared to clearIfaceCounters
	req := &vpe.CliInband{Cmd: "clear runtime"}
	reply := &vpe.CliInbandReply{}

	if err := s.apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return fmt.Errorf("request failed: %v", err)
	}

	return nil
}

// ClearErrorCounters clears the counters for errors.
func (s *VPP) ClearErrorCounters() error {
	// stats-api does not correctly pickup cleared error counters,
	// it was reported as a bug. You can find it here -> https://jira.fd.io/projects/GOVPP/issues/GOVPP-19?filter=allopenissues
	// TODO: implement me

	req := &vpe.CliInband{Cmd: "clear errors"}
	reply := &vpe.CliInbandReply{}

	if err := s.apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return fmt.Errorf("request failed: %v", err)
	}

	return nil
}

// Memory returns memory usage per thread.
func (s *VPP) Memory() ([]string, error) {
	req := &vpe.CliInband{Cmd: "show memory main-heap verbose"}
	reply := &vpe.CliInbandReply{}

	if err := s.apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	rows := make([]string, 0, 1) // there's gonna be at least 1 thread
	for _, r := range strings.Split(reply.Reply, "\n") {
		if r == "" {
			continue
		}
		rows = append(rows, strings.Trim(r, " \n"))
	}
	return rows, nil
}

// Threads returns thread data per thread.
func (s *VPP) Threads() ([]vpe.ThreadData, error) {
	req := &vpe.ShowThreads{}
	reply := &vpe.ShowThreadsReply{}

	if err := s.apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	return reply.ThreadData, nil
}
