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

package vpp

import (
	"context"
	"encoding/gob"
	"fmt"
	govppapi "git.fd.io/govpp.git/api"
	"github.com/PantheonTechnologies/vpptop/stats/api"
	"github.com/ligato/cn-infra/logging/logrus"
	"go.ligato.io/vpp-agent/v2/plugins/vpp"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/binapi"

	govppcalls "go.ligato.io/vpp-agent/v2/plugins/govppmux/vppcalls"
	telemetrycalls "go.ligato.io/vpp-agent/v2/plugins/telemetry/vppcalls"
	ifplugincalls "go.ligato.io/vpp-agent/v2/plugins/vpp/ifplugin/vppcalls"

	vpe1904 "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp1904/vpe"
	vpe1908 "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp1908/vpe"
	vpe2001 "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001/vpe"
	vpe2001_324 "go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324/vpe"

	"go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp1904"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp1908"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/binapi/vpp2001_324"

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

// HandlerDef is a VPP handler definition. It is used to validate
// compatibility with the version of the connected VPP
type HandlerDef struct{}

func (d *HandlerDef) IsHandlerCompatible(c *api.VppClient, isRemote bool) (api.HandlerAPI, bool, error) {
	ch, err := c.NewAPIChannel()
	if err != nil {
		return nil, false, err
	}
	binapiVersion, err := vpp.FindCompatibleBinapi(ch)
	if err == nil {
		return NewVPPHandler(c, ch, string(binapiVersion), isRemote), true, nil
	}
	return nil, false, nil
}

// Handler uses Ligato VPP-Agent interface and telemetry low-level handlers
// to obtain data from VPP
type Handler struct {
	vppCoreCalls      govppcalls.VppCoreAPI
	interfaceVppCalls ifplugincalls.InterfaceVppAPI
	telemetryVppCalls telemetrycalls.TelemetryVppAPI

	apiChan       govppapi.Channel
	binapiVersion string
}

// NewVPPHandler creates a new instance of the VPP Handler
func NewVPPHandler(c vpp.Client, ch govppapi.Channel, binapiVersion string, isRemote bool) *Handler {
	if isRemote {
		msgList := binapi.Versions[binapi.Version(binapiVersion)]
		for _, msg := range msgList.AllMessages() {
			gob.Register(msg)
		}
	}
	return &Handler{
		vppCoreCalls:      govppcalls.CompatibleHandler(c),
		interfaceVppCalls: ifplugincalls.CompatibleInterfaceVppHandler(c, logrus.NewLogger("")),
		telemetryVppCalls: telemetrycalls.CompatibleTelemetryHandler(c),
		binapiVersion:     binapiVersion,
		apiChan:           ch,
	}
}

func (h *Handler) RunCli(ctx context.Context, cmd string) (string, error) {
	return h.vppCoreCalls.RunCli(ctx, cmd)
}

func (h *Handler) DumpInterfaces(ctx context.Context) (map[uint32]*api.InterfaceDetails, error) {
	interfaceDetails := make(map[uint32]*api.InterfaceDetails)
	interfaceMap, err := h.interfaceVppCalls.DumpInterfaces(ctx)
	if err != nil {
		return nil, err
	}
	for swIfIdx, ifData := range interfaceMap {
		interfaceDetails[swIfIdx] = &api.InterfaceDetails{
			Name:         ifData.Interface.Name,
			InternalName: ifData.Meta.InternalName,
			SwIfIndex:    swIfIdx,
			IsEnabled:    ifData.Interface.Enabled,
			IPAddresses:  ifData.Interface.IpAddresses,
			MTU:          ifData.Meta.MTU,
		}
	}
	return interfaceDetails, nil
}

func (h *Handler) DumpInterfaceStats(ctx context.Context) (*govppapi.InterfaceStats, error) {
	return h.telemetryVppCalls.GetInterfaceStats(ctx)
}

func (h *Handler) DumpNodeCounters(ctx context.Context) (*api.NodeCounterInfo, error) {
	counters := make([]api.NodeCounter, 0)
	nodeCountersData, err := h.telemetryVppCalls.GetNodeCounters(ctx)
	if err != nil {
		return nil, err
	}
	for _, nodeCounter := range nodeCountersData.GetCounters() {
		counters = append(counters, api.NodeCounter{
			Value: nodeCounter.Value,
			Name:  nodeCounter.Name,
			Node:  nodeCounter.Node,
		})
	}
	return &api.NodeCounterInfo{
		Counters: counters,
	}, nil
}

func (h *Handler) DumpRuntimeInfo(ctx context.Context) (*api.RuntimeInfo, error) {
	threads := make([]api.RuntimeThread, 0)
	runtimeInfo, err := h.telemetryVppCalls.GetRuntimeInfo(ctx)
	if err != nil {
		return nil, err
	}
	for _, thread := range runtimeInfo.GetThreads() {
		items := make([]api.RuntimeItem, 0)
		for _, item := range thread.Items {
			items = append(items, api.RuntimeItem{
				Index:          item.Index,
				Name:           item.Name,
				State:          item.State,
				Calls:          item.Calls,
				Vectors:        item.Vectors,
				Suspends:       item.Suspends,
				Clocks:         item.Clocks,
				VectorsPerCall: item.VectorsPerCall,
			})
		}
		threads = append(threads, api.RuntimeThread{
			ID:                  thread.ID,
			Name:                thread.Name,
			Time:                thread.Time,
			AvgVectorsPerNode:   thread.AvgVectorsPerNode,
			LastMainLoops:       thread.LastMainLoops,
			VectorsPerMainLoop:  thread.VectorsPerMainLoop,
			VectorLengthPerNode: thread.VectorLengthPerNode,
			VectorRatesIn:       thread.VectorRatesIn,
			VectorRatesOut:      thread.VectorRatesOut,
			VectorRatesDrop:     thread.VectorRatesDrop,
			VectorRatesPunt:     thread.VectorRatesDrop,
			Items:               items,
		})
	}
	return &api.RuntimeInfo{
		Threads: threads,
	}, nil
}

func (h *Handler) DumpPlugins(ctx context.Context) ([]api.PluginInfo, error) {
	pluginInfo := make([]api.PluginInfo, 0)
	plugins, err := h.vppCoreCalls.GetPlugins(ctx)
	if err != nil {
		return nil, err
	}
	for _, plugin := range plugins {
		pluginInfo = append(pluginInfo, api.PluginInfo{
			Name:        plugin.Name,
			Path:        plugin.Path,
			Version:     plugin.Version,
			Description: plugin.Description,
		})
	}
	return pluginInfo, nil
}

func (h *Handler) DumpVersion(ctx context.Context) (*api.VersionInfo, error) {
	ver, err := h.vppCoreCalls.GetVersion(ctx)
	if err != nil {
		return nil, err
	}
	return &api.VersionInfo{
		Program:        ver.Program,
		Version:        ver.Version,
		BuildDate:      ver.BuildDate,
		BuildDirectory: ver.BuildDirectory,
	}, nil
}

func (h *Handler) DumpSession(ctx context.Context) (*api.SessionInfo, error) {
	session, err := h.vppCoreCalls.GetSession(ctx)
	if err != nil {
		return nil, err
	}
	return &api.SessionInfo{
		PID:       session.PID,
		ClientIdx: session.ClientIdx,
		Uptime:    session.Uptime,
	}, nil
}

func (h *Handler) DumpThreads(_ context.Context) ([]api.ThreadData, error) {
	switch h.binapiVersion {
	case vpp1904.Version:
		return h.threads1904()
	case vpp1908.Version:
		return h.threads1908()
	case vpp2001.Version:
		return h.threads2001()
	case vpp2001_324.Version:
		return h.threads2001324()
	default:
		return nil, fmt.Errorf("unsuported vpp version %v", h.binapiVersion)
	}
}

func (h *Handler) Close() {
	if h.apiChan != nil {
		h.apiChan.Close()
	}
}

func (h *Handler) threads1904() ([]api.ThreadData, error) {
	req := &vpe1904.ShowThreads{}
	reply := &vpe1904.ShowThreadsReply{}
	if err := h.apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	result := make([]api.ThreadData, len(reply.ThreadData))
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

func (h *Handler) threads1908() ([]api.ThreadData, error) {
	req := &vpe1908.ShowThreads{}
	reply := &vpe1908.ShowThreadsReply{}
	if err := h.apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	result := make([]api.ThreadData, len(reply.ThreadData))
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

func (h *Handler) threads2001() ([]api.ThreadData, error) {
	req := &vpe2001.ShowThreads{}
	reply := &vpe2001.ShowThreadsReply{}
	if err := h.apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	result := make([]api.ThreadData, len(reply.ThreadData))
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

func (h *Handler) threads2001324() ([]api.ThreadData, error) {
	req := &vpe2001_324.ShowThreads{}
	reply := &vpe2001_324.ShowThreadsReply{}
	if err := h.apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	result := make([]api.ThreadData, len(reply.ThreadData))
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
