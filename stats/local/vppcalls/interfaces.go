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

package vppcalls

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/PantheonTechnologies/vpptop/stats/api"
	dhcpapi "github.com/PantheonTechnologies/vpptop/stats/local/binapi/dhcp"
	interfaces "github.com/PantheonTechnologies/vpptop/stats/local/binapi/interface"
	"github.com/PantheonTechnologies/vpptop/stats/local/binapi/interface_types"
	"github.com/PantheonTechnologies/vpptop/stats/local/binapi/ip"
	"github.com/PantheonTechnologies/vpptop/stats/local/binapi/ip_types"
)

// InterfaceVppAPI defines interface-specific methods
type InterfaceVppAPI interface {
	DumpInterfaces(ctx context.Context) (map[uint32]*api.InterfaceDetails, error)
}

// InterfaceHandler implements InterfaceVppAPI
type InterfaceHandler struct {
	ch govppapi.Channel
}

// NewInterfaceHandler returns a new instance of the InterfaceVppAPI
func NewInterfaceHandler(ch govppapi.Channel) InterfaceVppAPI {
	h := &InterfaceHandler{
		ch: ch,
	}
	return h
}

const (
	// represents interface index of 'all'
	allInterfaces = ^uint32(0)
)

// Dhcp is helper struct for DHCP metadata, split to client and lease (similar to VPP binary API)
type dhcp struct {
	Client *client `json:"dhcp_client"`
	Lease  *lease  `json:"dhcp_lease"`
}

// Client is helper struct grouping DHCP client data
type client struct {
	SwIfIndex        uint32
	Hostname         string
	ID               string
	WantDhcpEvent    bool
	SetBroadcastFlag bool
	PID              uint32
}

// Lease is helper struct grouping DHCP lease data
type lease struct {
	SwIfIndex     uint32
	State         uint8
	Hostname      string
	IsIPv6        bool
	MaskWidth     uint8
	HostAddress   string
	RouterAddress string
	HostMac       string
}

// DumpInterfaces is simplified implementation retrieving only essential data for the VPPTop
func (h *InterfaceHandler) DumpInterfaces(_ context.Context) (map[uint32]*api.InterfaceDetails, error) {
	ifs, err := h.dumpInterfaces()
	if err != nil {
		return nil, err
	}
	// Retrieve DHCP clients (required for IP addresses)
	dhcpClients, err := h.dumpDhcpClients()
	if err != nil {
		return nil, fmt.Errorf("failed to dump interface DHCP clients: %v", err)
	}
	// Retrieve IP addresses
	err = h.dumpIPAddressDetails(ifs, false, dhcpClients)
	if err != nil {
		return nil, err
	}
	err = h.dumpIPAddressDetails(ifs, true, dhcpClients)
	if err != nil {
		return nil, err
	}

	return ifs, nil
}

func (h *InterfaceHandler) dumpInterfaces(ifIdxs ...uint32) (map[uint32]*api.InterfaceDetails, error) {
	ifs := make(map[uint32]*api.InterfaceDetails)

	ifIdx := allInterfaces
	if len(ifIdxs) > 0 {
		ifIdx = ifIdxs[0]
	}
	// All interfaces
	reqCtx := h.ch.SendMultiRequest(&interfaces.SwInterfaceDump{
		SwIfIndex: interface_types.InterfaceIndex(ifIdx),
	})
	for {
		ifDetails := &interfaces.SwInterfaceDetails{}
		stop, err := reqCtx.ReceiveReply(ifDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to dump interface: %v", err)
		}

		name := strings.TrimRight(ifDetails.InterfaceName, "\x00")
		details := &api.InterfaceDetails{
			Name:         strings.TrimRight(ifDetails.Tag, "\x00"),
			IsEnabled:    ifDetails.Flags&interface_types.IF_STATUS_API_FLAG_ADMIN_UP != 0,
			InternalName: name,
			SwIfIndex:    uint32(ifDetails.SwIfIndex),
			MTU:          ifDetails.Mtu,
		}
		ifs[uint32(ifDetails.SwIfIndex)] = details
	}

	return ifs, nil
}

// DumpDhcpClients returns a slice of DhcpMeta with all interfaces and other DHCP-related information available
func (h *InterfaceHandler) dumpDhcpClients() (map[uint32]*dhcp, error) {
	dhcpData := make(map[uint32]*dhcp)
	reqCtx := h.ch.SendMultiRequest(&dhcpapi.DHCPClientDump{})

	for {
		dhcpDetails := &dhcpapi.DHCPClientDetails{}
		last, err := reqCtx.ReceiveReply(dhcpDetails)
		if last {
			break
		}
		if err != nil {
			return nil, err
		}
		clientData := dhcpDetails.Client
		leaseData := dhcpDetails.Lease

		// DHCP client data
		dhcpClient := &client{
			SwIfIndex:        uint32(clientData.SwIfIndex),
			Hostname:         strings.TrimRight(clientData.Hostname, "\x00"),
			ID:               string(bytes.SplitN(clientData.ID, []byte{0x00}, 2)[0]),
			WantDhcpEvent:    clientData.WantDHCPEvent,
			SetBroadcastFlag: clientData.SetBroadcastFlag,
			PID:              clientData.PID,
		}

		// DHCP lease data
		dhcpLease := &lease{
			SwIfIndex:     uint32(leaseData.SwIfIndex),
			State:         uint8(leaseData.State),
			Hostname:      strings.TrimRight(leaseData.Hostname, "\x00"),
			IsIPv6:        leaseData.IsIPv6,
			HostAddress:   dhcpAddressToString(leaseData.HostAddress, uint32(leaseData.MaskWidth), leaseData.IsIPv6),
			RouterAddress: dhcpAddressToString(leaseData.RouterAddress, uint32(leaseData.MaskWidth), leaseData.IsIPv6),
			HostMac:       net.HardwareAddr(leaseData.HostMac[:]).String(),
		}

		// DHCP metadata
		dhcpData[uint32(clientData.SwIfIndex)] = &dhcp{
			Client: dhcpClient,
			Lease:  dhcpLease,
		}
	}

	return dhcpData, nil
}

// dumpIPAddressDetails dumps IP address details of interfaces from VPP and fills them into the provided interface map.
func (h *InterfaceHandler) dumpIPAddressDetails(ifs map[uint32]*api.InterfaceDetails, isIPv6 bool, dhcpClients map[uint32]*dhcp) error {
	// Dump IP addresses of each interface.
	for idx := range ifs {
		reqCtx := h.ch.SendMultiRequest(&ip.IPAddressDump{
			SwIfIndex: interface_types.InterfaceIndex(idx),
			IsIPv6:    isIPv6,
		})
		for {
			ipDetails := &ip.IPAddressDetails{}
			stop, err := reqCtx.ReceiveReply(ipDetails)
			if stop {
				break // Break from the loop.
			}
			if err != nil {
				return fmt.Errorf("failed to dump interface %d IP address details: %v", idx, err)
			}
			h.processIPDetails(ifs, ipDetails, dhcpClients)
		}
	}

	return nil
}

// processIPDetails processes ip.IPAddressDetails binary API message and fills the details into the provided interface map.
func (h *InterfaceHandler) processIPDetails(ifs map[uint32]*api.InterfaceDetails, ipDetails *ip.IPAddressDetails, dhcpClients map[uint32]*dhcp) {
	ifDetails, ifIdxExists := ifs[uint32(ipDetails.SwIfIndex)]
	if !ifIdxExists {
		return
	}

	var ipAddr string
	ipByte := make([]byte, 16)
	copy(ipByte[:], ipDetails.Prefix.Address.Un.XXX_UnionData[:])
	if ipDetails.Prefix.Address.Af == ip_types.ADDRESS_IP6 {
		ipAddr = fmt.Sprintf("%s/%d", net.IP(ipByte).To16().String(), uint32(ipDetails.Prefix.Len))
	} else {
		ipAddr = fmt.Sprintf("%s/%d", net.IP(ipByte[:4]).To4().String(), uint32(ipDetails.Prefix.Len))
	}

	// skip IP addresses given by DHCP
	if dhcpClient, hasDhcpClient := dhcpClients[uint32(ipDetails.SwIfIndex)]; hasDhcpClient {
		if dhcpClient.Lease != nil && dhcpClient.Lease.HostAddress == ipAddr {
			return
		}
	}

	ifDetails.IPAddresses = append(ifDetails.IPAddresses, ipAddr)
}

func dhcpAddressToString(address ip_types.Address, maskWidth uint32, isIPv6 bool) string {
	dhcpIPByte := make([]byte, 16)
	copy(dhcpIPByte[:], address.Un.XXX_UnionData[:])
	if isIPv6 {
		return fmt.Sprintf("%s/%d", net.IP(dhcpIPByte).To16().String(), maskWidth)
	}
	return fmt.Sprintf("%s/%d", net.IP(dhcpIPByte[:4]).To4().String(), maskWidth)
}
