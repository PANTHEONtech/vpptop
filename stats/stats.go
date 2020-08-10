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
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/adapter/statsclient"
	"github.com/PantheonTechnologies/vpptop/local"
	"strings"
	"sync"
	"time"

	"git.fd.io/govpp.git/adapter"
	"git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
	"git.fd.io/govpp.git/proxy"

	"go.ligato.io/vpp-agent/v2/plugins/govppmux"
	govppcalls "go.ligato.io/vpp-agent/v2/plugins/govppmux/vppcalls"
	telemetrycalls "go.ligato.io/vpp-agent/v2/plugins/telemetry/vppcalls"
	ifplugincalls "go.ligato.io/vpp-agent/v2/plugins/vpp/ifplugin/vppcalls"

	"go.ligato.io/vpp-agent/v2/plugins/vpp"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/binapi"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp1904"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp1908"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324"

	vpe_local "github.com/PantheonTechnologies/vpptop/local/binapi/vpe"
	vpe1904 "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp1904/vpe"
	vpe1908 "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp1908/vpe"
	vpe2001 "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001/vpe"
	vpe2001_324 "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/vpe"

	// import for handler ifplugin handler registration
	_ "go.ligato.io/vpp-agent/v2/plugins/vpp/ifplugin/vppcalls/vpp1904"
	_ "go.ligato.io/vpp-agent/v2/plugins/vpp/ifplugin/vppcalls/vpp1908"
	_ "go.ligato.io/vpp-agent/v2/plugins/vpp/ifplugin/vppcalls/vpp2001"
	_ "go.ligato.io/vpp-agent/v2/plugins/vpp/ifplugin/vppcalls/vpp2001_324"

	// import for handler telemetry handler registration
	_ "go.ligato.io/vpp-agent/v2/plugins/telemetry/vppcalls/vpp1904"
	_ "go.ligato.io/vpp-agent/v2/plugins/telemetry/vppcalls/vpp1908"
	_ "go.ligato.io/vpp-agent/v2/plugins/telemetry/vppcalls/vpp2001"
	_ "go.ligato.io/vpp-agent/v2/plugins/telemetry/vppcalls/vpp2001_324"
)

const (
	stateUp   = "up"
	stateDown = "down"
)

// VPPTopStatsHandler defines methods required to obtain all stats displayed by the VPPTop.
// It effectively replaces specific plugin-based handlers (interface, telemetry... )
type VPPTopStatsHandler interface {
	// GetPlugins retrieves info about loaded VPP plugins.
	GetPlugins(context.Context) ([]govppcalls.PluginInfo, error)

	// RunCli sends CLI command to VPP
	RunCli(ctx context.Context, cmd string) (string, error)

	// GetVersion retrieves info about VPP version.
	GetVersion(context.Context) (*govppcalls.VersionInfo, error)

	// GetSession retrieves info about active session
	GetSession(context.Context) (*govppcalls.SessionInfo, error)

	// DumpInterfaces retrieves VPP interface data and returns them as VPP-Agent northbound
	// interface data
	DumpInterfaces(ctx context.Context) (map[uint32]*ifplugincalls.InterfaceDetails, error)

	// GetInterfaceStats retrieves interface stats
	GetInterfaceStats(context.Context) (*api.InterfaceStats, error)

	// GetNodeCounters retrieves node counters info
	GetNodeCounters(context.Context) (*telemetrycalls.NodeCounterInfo, error)

	// GetRuntimeInfo retrieves node's runtime info
	GetRuntimeInfo(context.Context) (*telemetrycalls.RuntimeInfo, error)
}

type (
	vppClient struct {
		apiChan   api.Channel
		statsConn api.StatsProvider
		vppInfo   govppmux.VPPInfo
		client    *proxy.Client
		vppConn   *core.Connection
	}

	// VPP provides statistics about vpp
	// such as runtime counters, interface counters,
	// error counters...
	VPP struct {
		vppclient   *vppClient
		statsClient adapter.StatsAPI

		handler VPPTopStatsHandler

		binapiVersion     binapi.Version
		vppVersion        *govppcalls.VersionInfo
		lastErrorCounters map[string]uint64
	}

	// Interface wraps all interface counters.
	Interface struct {
		api.InterfaceCounters
		IPAddrs []string
		State   string
		MTU     []uint32
	}

	// ThreadData wraps all thread data counters.
	ThreadData struct {
		ID        uint32
		Name      string
		Type      string
		PID       uint32
		CPUID     uint32
		Core      uint32
		CPUSocket uint32
	}

	// Node wraps all node counters.
	Node telemetrycalls.RuntimeItem

	// Error wraps all error counters.
	Error telemetrycalls.NodeCounter

	// Memory wraps all memory counters.
	Memory telemetrycalls.MemoryThread
)

func (c *vppClient) NewAPIChannel() (api.Channel, error) {
	if c.client != nil {
		return c.client.NewBinapiClient()
	}

	return c.vppConn.NewAPIChannel()
}

func (c *vppClient) Stats() api.StatsProvider {
	return c.statsConn
}

func (c *vppClient) CheckCompatiblity(msgs ...api.Message) error {
	if c.apiChan == nil {
		ch, err := c.NewAPIChannel()
		if err != nil {
			return err
		}
		c.apiChan = ch
	}
	return c.apiChan.CheckCompatiblity(msgs...)
}

func (c *vppClient) IsPluginLoaded(plugin string) bool {
	for _, p := range c.vppInfo.Plugins {
		if p.Name == plugin {
			return true
		}
	}
	return false
}

// ConnectRemote connects vpptop to a remote proxy providing vpp statistics.
func (s *VPP) ConnectRemote(raddr string) error {
	s.lastErrorCounters = make(map[string]uint64)

	var err error
	var client *proxy.Client
	for i := 0; i < 3; i++ {
		client, err = proxy.Connect(raddr)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to raddr %v, reason: %v", raddr, err)
	}

	statsConn, err := client.NewStatsClient()
	if err != nil {
		return err
	}

	apiChan, err := client.NewBinapiClient()
	if err != nil {
		return err
	}

	s.vppclient = &vppClient{
		client:    client,
		statsConn: statsConn,
		apiChan:   apiChan,
	}

	var useLocal bool
	if s.binapiVersion, useLocal = CheckLocalHandlerCompatible(s.vppclient.apiChan); !useLocal {
		s.binapiVersion, err = vpp.FindCompatibleBinapi(s.vppclient.apiChan)
		if err != nil {
			return err
		}
	}

	msgList := binapi.Versions[s.binapiVersion]
	for _, msg := range msgList.AllMessages() {
		gob.Register(msg)
	}

	if useLocal {
		s.handler = NewCompatibleLocalHandler(s.vppclient)
	} else {
		s.handler = NewCompatibleVPPHandler(s.vppclient)
	}

	ctx := context.Background()

	plugins, err := s.handler.GetPlugins(ctx)
	if err != nil {
		return err
	}

	session, err := s.handler.GetSession(ctx)
	if err != nil {
		return err
	}

	s.vppVersion, err = s.handler.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get vpp version: %v", err)
	}

	s.vppclient.vppInfo = govppmux.VPPInfo{
		Connected:   true,
		VersionInfo: *s.vppVersion,
		SessionInfo: *session,
		Plugins:     plugins,
	}

	return nil
}

// Connect establishes a connection to govpp API.
func (s *VPP) Connect(soc string) error {
	s.lastErrorCounters = make(map[string]uint64)

	s.statsClient = statsclient.NewStatsClient(soc)

	statsConn, err := core.ConnectStats(s.statsClient)
	if err != nil {
		return fmt.Errorf("connection to stats api failed: %v", err)
	}

	vppConn, err := govpp.Connect("")
	if err != nil {
		return fmt.Errorf("connection to govpp failed: %v", err)
	}

	apiChan, err := vppConn.NewAPIChannel()
	if err != nil {
		return err
	}

	s.vppclient = &vppClient{
		vppConn:   vppConn,
		statsConn: statsConn,
		apiChan:   apiChan,
	}

	var useLocal bool
	if s.binapiVersion, useLocal = CheckLocalHandlerCompatible(s.vppclient.apiChan); !useLocal {
		s.binapiVersion, err = vpp.FindCompatibleBinapi(s.vppclient.apiChan)
		if err != nil {
			return err
		}
	}

	if useLocal {
		s.handler = NewCompatibleLocalHandler(s.vppclient)
	} else {
		s.handler = NewCompatibleVPPHandler(s.vppclient)
	}

	ctx := context.Background()
	plugins, err := s.handler.GetPlugins(ctx)
	if err != nil {
		return err
	}

	session, err := s.handler.GetSession(ctx)
	if err != nil {
		return err
	}

	s.vppVersion, err = s.handler.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get vpp version: %v", err)
	}

	s.vppclient.vppInfo = govppmux.VPPInfo{
		Connected:   true,
		VersionInfo: *s.vppVersion,
		SessionInfo: *session,
		Plugins:     plugins,
	}

	return nil
}

// Version returns the current vpp version.
func (s *VPP) Version() (string, error) {
	return "VPP version: " + s.vppVersion.Version + "\n" + s.vppVersion.BuildDate, nil
}

// Disconnect should be called after Connect, if the connection is no longer needed.
func (s *VPP) Disconnect() {
	if s.vppclient != nil {
		if s.vppclient.vppConn != nil {
			s.vppclient.vppConn.Disconnect()
		}
		if s.vppclient.apiChan != nil {
			s.vppclient.apiChan.Close()
		}
	}

	if s.statsClient != nil {
		s.statsClient.Disconnect()
	}
}

// GetNodes returns per node statistics.
func (s *VPP) GetNodes(ctx context.Context) ([]Node, error) {
	runtimeCounters, err := s.handler.GetRuntimeInfo(ctx)
	if err != nil {
		return nil, err
	}

	threads := runtimeCounters.GetThreads()
	if len(threads) == 0 {
		return nil, errors.New("no runtime counters")
	}

	result := make([]Node, 0, len(threads[0].Items))
	for _, thread := range threads {
		for _, item := range thread.Items {
			result = append(result, Node(item))
		}
	}

	return result, nil
}

// GetInterfaces returns per interface statistics.
func (s *VPP) GetInterfaces(ctx context.Context) ([]Interface, error) {
	var ifaceStats *api.InterfaceStats
	var ifaceDetails map[uint32]*ifplugincalls.InterfaceDetails

	wg := new(sync.WaitGroup)
	wg.Add(2)

	errChan := make(chan error, 2)

	go func() {
		wg.Wait()
		close(errChan)
	}()

	go func() {
		defer wg.Done()
		var err error
		ifaceDetails, err = s.handler.DumpInterfaces(ctx)
		errChan <- err
	}()

	go func() {
		defer wg.Done()
		var err error
		ifaceStats, err = s.handler.GetInterfaceStats(ctx)
		errChan <- err
	}()

	for err := range errChan {
		if err != nil {
			return nil, fmt.Errorf("request failed: %v", err)
		}
	}

	result := make([]Interface, 0, len(ifaceDetails))
	for _, iface := range ifaceStats.Interfaces {
		details, ok := ifaceDetails[iface.InterfaceIndex]
		if !ok {
			continue
		}
		state := stateDown
		if details.Interface.GetEnabled() {
			state = stateUp
		}
		result = append(result, Interface{
			InterfaceCounters: iface,
			IPAddrs:           details.Interface.GetIpAddresses(),
			State:             state,
			MTU:               details.Meta.MTU,
		})
	}
	return result, nil
}

// GetErrors returns per error statistics.
func (s *VPP) GetErrors(ctx context.Context) ([]Error, error) {
	counters, err := s.handler.GetNodeCounters(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Error, 0)
	for _, counter := range counters.GetCounters() {
		counter.Value -= s.lastErrorCounters[counter.Node+counter.Name]
		if counter.Value == 0 {
			continue
		}
		result = append(result, Error(counter))
	}

	return result, nil
}

// ClearIfaceCounters resets the counters for the interface.
func (s *VPP) ClearIfaceCounters(ctx context.Context) error {
	if _, err := s.handler.RunCli(ctx, "clear interfaces"); err != nil {
		return fmt.Errorf("request failed: %v", err)
	}

	return nil
}

// ClearRuntimeCounters clears the runtime counters for nodes.
func (s *VPP) ClearRuntimeCounters(ctx context.Context) error {
	if _, err := s.handler.RunCli(ctx, "clear runtime"); err != nil {
		return fmt.Errorf("request failed: %v", err)
	}

	return nil
}

// ClearErrorCounters clears the counters for errors.
func (s *VPP) ClearErrorCounters(ctx context.Context) error {
	s.updateLastErrors(ctx)
	if _, err := s.handler.RunCli(ctx, "clear errors"); err != nil {
		return fmt.Errorf("request failed: %v", err)
	}

	return nil
}

// Memory returns memory usage per thread.
func (s *VPP) Memory(ctx context.Context) ([]string, error) {
	mem, err := s.handler.RunCli(ctx, "show memory main-heap verbose")
	if err != nil {
		return nil, err
	}

	rows := make([]string, 0, 1) // there's gonna be at least 1 thread
	for _, r := range strings.Split(mem, "\n") {
		if r == "" {
			continue
		}

		rows = append(rows, strings.Trim(r, " \n"))
	}

	return rows, nil
}

// Threads returns thread data per thread.
func (s *VPP) Threads(_ context.Context) ([]ThreadData, error) {
	switch s.binapiVersion {
	case local.Version:
		return s.threadsLocal()
	case vpp1904.Version:
		return s.threads1904()
	case vpp1908.Version:
		return s.threads1908()
	case vpp2001.Version:
		return s.threads2001()
	case vpp2001_324.Version:
		return s.threads2001324()
	default:
		return nil, fmt.Errorf("unsuported vpp version %v", s.binapiVersion)
	}
}

func (s *VPP) threadsLocal() ([]ThreadData, error) {
	req := &vpe_local.ShowThreads{}
	reply := &vpe_local.ShowThreadsReply{}
	if err := s.vppclient.apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	result := make([]ThreadData, len(reply.ThreadData))
	for i := range reply.ThreadData {
		result[i].ID = reply.ThreadData[i].ID
		result[i].Name = reply.ThreadData[i].Name
		result[i].Type = reply.ThreadData[i].Type
		result[i].PID = reply.ThreadData[i].PID
		result[i].Core = reply.ThreadData[i].Core
		result[i].CPUID = reply.ThreadData[i].CPUID
		result[i].CPUSocket = reply.ThreadData[i].CPUSocket
	}

	return result, nil
}

func (s *VPP) threads1904() ([]ThreadData, error) {
	req := &vpe1904.ShowThreads{}
	reply := &vpe1904.ShowThreadsReply{}
	if err := s.vppclient.apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	result := make([]ThreadData, len(reply.ThreadData))
	for i := range reply.ThreadData {
		result[i].ID = reply.ThreadData[i].ID
		result[i].Name = string(reply.ThreadData[i].Name)
		result[i].Type = string(reply.ThreadData[i].Type)
		result[i].PID = reply.ThreadData[i].PID
		result[i].Core = reply.ThreadData[i].Core
		result[i].CPUID = reply.ThreadData[i].CPUID
		result[i].CPUSocket = reply.ThreadData[i].CPUSocket
	}

	return result, nil
}

func (s *VPP) threads1908() ([]ThreadData, error) {
	req := &vpe1908.ShowThreads{}
	reply := &vpe1908.ShowThreadsReply{}
	if err := s.vppclient.apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	result := make([]ThreadData, len(reply.ThreadData))
	for i := range reply.ThreadData {
		result[i].ID = reply.ThreadData[i].ID
		result[i].Name = string(reply.ThreadData[i].Name)
		result[i].Type = string(reply.ThreadData[i].Type)
		result[i].PID = reply.ThreadData[i].PID
		result[i].Core = reply.ThreadData[i].Core
		result[i].CPUID = reply.ThreadData[i].CPUID
		result[i].CPUSocket = reply.ThreadData[i].CPUSocket
	}

	return result, nil
}

func (s *VPP) threads2001324() ([]ThreadData, error) {
	req := &vpe2001_324.ShowThreads{}
	reply := &vpe2001_324.ShowThreadsReply{}
	if err := s.vppclient.apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	result := make([]ThreadData, len(reply.ThreadData))
	for i := range reply.ThreadData {
		result[i].ID = reply.ThreadData[i].ID
		result[i].Name = string(reply.ThreadData[i].Name)
		result[i].Type = string(reply.ThreadData[i].Type)
		result[i].PID = reply.ThreadData[i].PID
		result[i].Core = reply.ThreadData[i].Core
		result[i].CPUID = reply.ThreadData[i].CPUID
		result[i].CPUSocket = reply.ThreadData[i].CPUSocket
	}

	return result, nil
}

func (s *VPP) threads2001() ([]ThreadData, error) {
	req := &vpe2001.ShowThreads{}
	reply := &vpe2001.ShowThreadsReply{}
	if err := s.vppclient.apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	result := make([]ThreadData, len(reply.ThreadData))
	for i := range reply.ThreadData {
		result[i].ID = reply.ThreadData[i].ID
		result[i].Name = string(reply.ThreadData[i].Name)
		result[i].Type = string(reply.ThreadData[i].Type)
		result[i].PID = reply.ThreadData[i].PID
		result[i].Core = reply.ThreadData[i].Core
		result[i].CPUID = reply.ThreadData[i].CPUID
		result[i].CPUSocket = reply.ThreadData[i].CPUSocket
	}

	return result, nil
}

// updateLastErrors clears the error counters.
func (s *VPP) updateLastErrors(ctx context.Context) {
	counters, err := s.handler.GetNodeCounters(ctx)
	if err != nil {
		return
	}

	for _, counter := range counters.GetCounters() {
		if counter.Value == 0 {
			continue
		}
		s.lastErrorCounters[counter.Node+counter.Name] = counter.Value
	}
}
