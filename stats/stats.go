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
	"git.fd.io/govpp.git/adapter/vppapiclient"
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
	DefaultSocket = vppapiclient.DefaultStatSocket

	client    adapter.StatsAPI
	statsConn *core.StatsConnection

	vppConn *core.Connection
	apiChan api.Channel
)

// Nodes extends the counters from the api.NodeCounters
// to also include Vectors/Calls.
type Nodes struct {
	api.NodeCounters
	VC float64
}

// Interfaces extends the counters from the api.InterfaceCounters
// to also include Ipv4 addresses, IPv6 addresses, state and Mtu.
type Interfaces struct {
	api.InterfaceCounters
	IPv4  []string
	IPv6  []string
	State string
	Mtu   []uint32
}

// Errors counters.
type Errors struct {
	Value    uint64
	NodeName string
	Reason   string
}

// Connect establishes a connection to the govpp API.
func Connect(soc string) error {
	client = vppapiclient.NewStatClient(soc)

	var err error
	statsConn, err = core.ConnectStats(client)
	if err != nil {
		return err
	}

	vppConn, err = govpp.Connect("")
	if err != nil {
		return err
	}

	apiChan, err = vppConn.NewAPIChannel()
	if err != nil {
		return err
	}

	var msgs []api.Message
	msgs = append(msgs, interfaces.Messages...)
	msgs = append(msgs, ip.Messages...)
	msgs = append(msgs, vpe.Messages...)
	err = apiChan.CheckCompatiblity(msgs...)
	if err != nil {
		return fmt.Errorf("compatibility check failed: %v", err)
	}

	return nil
}

// Version returns the current vpp version.
func Version() (string, error) {
	req := &vpe.ShowVersion{}
	reply := &vpe.ShowVersionReply{}

	if err := apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return "", err
	}
	return "VPP version: " + reply.Version + "\n" + reply.BuildDate, nil
}

// Disconnect should be called after Connect, if the connection is no longer needed.
func Disconnect() {
	apiChan.Close()
	vppConn.Disconnect()
	statsConn.Disconnect()
}

// interfacesDetails returns the details of each interface from the bin_api.
func interfacesDetails() (map[uint32]*interfaces.SwInterfaceDetails, error) {
	result := make(map[uint32]*interfaces.SwInterfaceDetails)

	req := apiChan.SendMultiRequest(&interfaces.SwInterfaceDump{})

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
func ipv4Addresses(ifIndex uint32) ([]string, error) {
	var ipv4Addresses []string
	req := &ip.IPAddressDump{SwIfIndex: ifIndex, IsIPv6: 0}
	reqCtx := apiChan.SendMultiRequest(req)
	for {
		ipDetails := &ip.IPAddressDetails{}
		stop, err := reqCtx.ReceiveReply(ipDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		ipv4Addresses = append(ipv4Addresses, net.IP(ipDetails.IP[:4]).String())
	}

	return ipv4Addresses, nil
}

// ipv6Addresses returns all IPv6 addresses assigned to the interface.
func ipv6Addresses(ifIndex uint32) ([]string, error) {
	var ipv6Addresses []string
	req := &ip.IPAddressDump{SwIfIndex: ifIndex, IsIPv6: 1}
	reqCtx := apiChan.SendMultiRequest(req)
	for {
		ipDetails := &ip.IPAddressDetails{}
		stop, err := reqCtx.ReceiveReply(ipDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		ipv6Addresses = append(ipv6Addresses, net.IP(ipDetails.IP).String())
	}
	return ipv6Addresses, nil
}

// GetNodes returns per node statistics.
func GetNodes() ([]Nodes, error) {
	var result []Nodes

	stats, err := statsConn.GetNodeStats()
	if err != nil {
		return nil, err
	}

	result = make([]Nodes, len(stats.Nodes))
	for i, node := range stats.Nodes {
		vc := 0.0
		if node.Vectors != 0 && node.Calls != 0 {
			vc = float64(node.Vectors) / float64(node.Calls)
		}
		result[i] = Nodes{
			NodeCounters: node,
			VC:           vc,
		}
	}
	return result, nil
}

// GetInterfaces returns per interface statistics.
func GetInterfaces() ([]Interfaces, error) {
	var result []Interfaces

	stats, err := statsConn.GetInterfaceStats()
	if err != nil {
		return nil, err
	}

	ifaces, err := interfacesDetails()
	if err != nil {
		return nil, err
	}

	// stats-api for interfaces returns counters for deleted interfaces which
	// are no longer present via vppctl. Tt was reported as a bug.
	// You can find it here -> https://jira.fd.io/projects/GOVPP/issues/GOVPP-18?filter=allopenissues

	// To bypass this for now, we check if the bin_api contains the interface
	// from the stats-api.
	result = make([]Interfaces, 0, len(ifaces))
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
		ifaceIPv4, err := ipv4Addresses(iface.InterfaceIndex)
		if err != nil {
			ifaceIPv4 = make([]string, 0)
			//return nil, err
		}
		ifaceIPv6, err := ipv6Addresses(iface.InterfaceIndex)
		if err != nil {
			ifaceIPv6 = make([]string, 0)
			//return nil, err
		}

		result = append(result, Interfaces{
			InterfaceCounters: iface,
			//Name:              string(ifaceDetails.InterfaceName),
			Mtu:   append([]uint32(nil), ifaceDetails.Mtu...),
			IPv4:  append(make([]string, 0, len(ifaceIPv4)), ifaceIPv4...),
			IPv6:  append(make([]string, 0, len(ifaceIPv6)), ifaceIPv6...),
			State: ifaceState,
		})
	}
	return result, nil
}

// GetErrors returns per error statistics.
func GetErrors() ([]Errors, error) {
	var result []Errors

	stats, err := statsConn.GetErrorStats("")
	if err != nil {
		return nil, err
	}

	result = make([]Errors, 0)
	for _, counter := range stats.Errors {
		if counter.Value == 0 {
			continue
		}
		s := strings.Split(counter.CounterName, "/")
		result = append(result, Errors{
			Value:    counter.Value,
			NodeName: s[0],
			Reason:   s[1],
		})
	}
	return result, nil
}

// ClearIfaceCounters resets the counters for the interface.
func ClearIfaceCounters(ifaceIndex uint32) error {
	req := &interfaces.SwInterfaceClearStats{SwIfIndex: ifaceIndex}
	reply := &interfaces.SwInterfaceClearStatsReply{}

	if err := apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// ClearRuntimeCounters clears the runtime counters for nodes.
func ClearRuntimeCounters() error {
	// TODO: find if there exits a bin_api call for clearing node counters 'casue this seems to be slow compared to clearIfaceCounters
	req := &vpe.CliInband{Cmd: "clear runtime"}
	reply := &vpe.CliInbandReply{}

	if err := apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return err
	}

	return nil
}

// ClearErrorCounters clears the counters for errors.
func ClearErrorCounters() error {
	// stats-api does not correctly pickup cleared error counters,
	// it was reported as a bug. You can find it here -> https://jira.fd.io/projects/GOVPP/issues/GOVPP-19?filter=allopenissues
	// TODO: implement me

	//req := &vpe.CliInband{Cmd:"clear errors"}
	//reply := &vpe.CliInbandReply{}
	//
	//if err := apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
	//	return err
	//}

	return nil
}

// Memory returns memory usage per thread.
func Memory() ([]string, error) {
	req := &vpe.CliInband{Cmd: "show memory verbose"}
	reply := &vpe.CliInbandReply{}

	if err := apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
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
func Threads() ([]vpe.ThreadData, error) {
	req := &vpe.ShowThreads{}
	reply := &vpe.ShowThreadsReply{}

	if err := apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	}

	return reply.ThreadData, nil
}
