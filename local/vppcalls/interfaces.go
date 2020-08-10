package vppcalls

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"

	"git.fd.io/govpp.git/api"
	"github.com/PantheonTechnologies/vpptop/local/binapi/dhcp"
	"github.com/PantheonTechnologies/vpptop/local/binapi/interfaces"
	"github.com/PantheonTechnologies/vpptop/local/binapi/ip"
	ifplugincalls "go.ligato.io/vpp-agent/v2/plugins/vpp/ifplugin/vppcalls"
	ifaces "go.ligato.io/vpp-agent/v2/proto/ligato/vpp/interfaces"
)

// InterfaceVppAPI defines interface-specific methods
type InterfaceVppAPI interface {
	DumpInterfaces(ctx context.Context) (map[uint32]*ifplugincalls.InterfaceDetails, error)
}

// InterfaceHandler implements InterfaceVppAPI
type InterfaceHandler struct {
	ch api.Channel
}

// NewInterfaceHandler returns a new instance of the InterfaceVppAPI
func NewInterfaceHandler(ch api.Channel) InterfaceVppAPI {
	h := &InterfaceHandler{
		ch: ch,
	}
	return h
}

const (
	// represents interface index of 'all'
	allInterfaces = ^uint32(0)
)

// DumpInterfaces is simplified implementation retrieving only essential data for the VPPTop
func (h *InterfaceHandler) DumpInterfaces(_ context.Context) (map[uint32]*ifplugincalls.InterfaceDetails, error) {
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

func (h *InterfaceHandler) dumpInterfaces(ifIdxs ...uint32) (map[uint32]*ifplugincalls.InterfaceDetails, error) {
	ifs := make(map[uint32]*ifplugincalls.InterfaceDetails)

	ifIdx := allInterfaces
	if len(ifIdxs) > 0 {
		ifIdx = ifIdxs[0]
	}
	// All interfaces
	reqCtx := h.ch.SendMultiRequest(&interfaces.SwInterfaceDump{
		SwIfIndex: interfaces.InterfaceIndex(ifIdx),
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

		ifaceName := strings.TrimRight(ifDetails.InterfaceName, "\x00")
		details := &ifplugincalls.InterfaceDetails{
			Interface: &ifaces.Interface{
				Name:    strings.TrimRight(ifDetails.Tag, "\x00"),
				Enabled: ifDetails.Flags&interfaces.IF_STATUS_API_FLAG_ADMIN_UP != 0,
			},
			Meta: &ifplugincalls.InterfaceMeta{
				SwIfIndex:    uint32(ifDetails.SwIfIndex),
				InternalName: ifaceName,
				MTU:          ifDetails.Mtu,
			},
		}
		ifs[uint32(ifDetails.SwIfIndex)] = details
	}

	return ifs, nil
}

// DumpDhcpClients returns a slice of DhcpMeta with all interfaces and other DHCP-related information available
func (h *InterfaceHandler) dumpDhcpClients() (map[uint32]*ifplugincalls.Dhcp, error) {
	dhcpData := make(map[uint32]*ifplugincalls.Dhcp)
	reqCtx := h.ch.SendMultiRequest(&dhcp.DHCPClientDump{})

	for {
		dhcpDetails := &dhcp.DHCPClientDetails{}
		last, err := reqCtx.ReceiveReply(dhcpDetails)
		if last {
			break
		}
		if err != nil {
			return nil, err
		}
		client := dhcpDetails.Client
		lease := dhcpDetails.Lease

		// DHCP client data
		dhcpClient := &ifplugincalls.Client{
			SwIfIndex:        uint32(client.SwIfIndex),
			Hostname:         strings.TrimRight(client.Hostname, "\x00"),
			ID:               string(bytes.SplitN(client.ID, []byte{0x00}, 2)[0]),
			WantDhcpEvent:    client.WantDHCPEvent,
			SetBroadcastFlag: client.SetBroadcastFlag,
			PID:              client.PID,
		}

		// DHCP lease data
		dhcpLease := &ifplugincalls.Lease{
			SwIfIndex:     uint32(lease.SwIfIndex),
			State:         uint8(lease.State),
			Hostname:      strings.TrimRight(lease.Hostname, "\x00"),
			IsIPv6:        lease.IsIPv6,
			HostAddress:   dhcpAddressToString(lease.HostAddress, uint32(lease.MaskWidth), lease.IsIPv6),
			RouterAddress: dhcpAddressToString(lease.RouterAddress, uint32(lease.MaskWidth), lease.IsIPv6),
			HostMac:       net.HardwareAddr(lease.HostMac[:]).String(),
		}

		// DHCP metadata
		dhcpData[uint32(client.SwIfIndex)] = &ifplugincalls.Dhcp{
			Client: dhcpClient,
			Lease:  dhcpLease,
		}
	}

	return dhcpData, nil
}

// dumpIPAddressDetails dumps IP address details of interfaces from VPP and fills them into the provided interface map.
func (h *InterfaceHandler) dumpIPAddressDetails(ifs map[uint32]*ifplugincalls.InterfaceDetails, isIPv6 bool, dhcpClients map[uint32]*ifplugincalls.Dhcp) error {
	// Dump IP addresses of each interface.
	for idx := range ifs {
		reqCtx := h.ch.SendMultiRequest(&ip.IPAddressDump{
			SwIfIndex: ip.InterfaceIndex(idx),
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
func (h *InterfaceHandler) processIPDetails(ifs map[uint32]*ifplugincalls.InterfaceDetails, ipDetails *ip.IPAddressDetails, dhcpClients map[uint32]*ifplugincalls.Dhcp) {
	ifDetails, ifIdxExists := ifs[uint32(ipDetails.SwIfIndex)]
	if !ifIdxExists {
		return
	}

	var ipAddr string
	ipByte := make([]byte, 16)
	copy(ipByte[:], ipDetails.Prefix.Address.Un.XXX_UnionData[:])
	if ipDetails.Prefix.Address.Af == ip.ADDRESS_IP6 {
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

	ifDetails.Interface.IpAddresses = append(ifDetails.Interface.IpAddresses, ipAddr)
}

func dhcpAddressToString(address dhcp.Address, maskWidth uint32, isIPv6 bool) string {
	dhcpIPByte := make([]byte, 16)
	copy(dhcpIPByte[:], address.Un.XXX_UnionData[:])
	if isIPv6 {
		return fmt.Sprintf("%s/%d", net.IP(dhcpIPByte).To16().String(), maskWidth)
	}
	return fmt.Sprintf("%s/%d", net.IP(dhcpIPByte[:4]).To4().String(), maskWidth)
}
