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
	govppapi "git.fd.io/govpp.git/api"
	"github.com/PantheonTechnologies/vpptop/stats/api"
	"go.ligato.io/cn-infra/v2/logging/logrus"
	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi"

	govppcalls "go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls"
	telemetrycalls "go.ligato.io/vpp-agent/v3/plugins/telemetry/vppcalls"
	ifplugincalls "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"

	// import for handler ifplugin handler registration
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls/vpp1908"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls/vpp2001"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls/vpp2005"
	_ "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls/vpp2009"

	// import for handler telemetry handler registration
	_ "go.ligato.io/vpp-agent/v3/plugins/telemetry/vppcalls/vpp1908"
	_ "go.ligato.io/vpp-agent/v3/plugins/telemetry/vppcalls/vpp2001"
	_ "go.ligato.io/vpp-agent/v3/plugins/telemetry/vppcalls/vpp2005"
	_ "go.ligato.io/vpp-agent/v3/plugins/telemetry/vppcalls/vpp2009"
)

// HandlerDef is a VPP handler definition. It is used to validate
// compatibility with the version of the connected VPP
type HandlerDef struct{}

func (d *HandlerDef) IsHandlerCompatible(c *api.VppClient, isRemote bool) (api.HandlerAPI, string, error) {
	ch, err := c.NewAPIChannel()
	if err != nil {
		return nil, "", err
	}
	binapiVersion, err := binapi.CompatibleVersion(ch)
	if err == nil {
		return NewVPPHandler(c, ch, string(binapiVersion), isRemote), string(binapiVersion), nil
	}
	return nil, "", nil
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
			Count:  nodeCounter.Value,
			Node:   nodeCounter.Node,
			Reason: nodeCounter.Name,
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

func (h *Handler) DumpThreads(ctx context.Context) ([]api.ThreadData, error) {
	threads, err := h.telemetryVppCalls.GetThreads(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]api.ThreadData, len(threads.GetItems()))
	for i, thread := range threads.GetItems() {
		result[i].ID = thread.ID
		result[i].Name = thread.Name
		result[i].Type = thread.Type
		result[i].PID = thread.PID
		result[i].Core = thread.Core
		result[i].CPUID = thread.CPUID
		result[i].CPUSocket = thread.CPUSocket
	}

	return result, nil
}

func (h *Handler) Close() {
	if h.apiChan != nil {
		h.apiChan.Close()
	}
}