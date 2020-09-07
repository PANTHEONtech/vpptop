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

package local

import (
	"context"
	"encoding/gob"
	"fmt"
	govppapi "git.fd.io/govpp.git/api"
	"github.com/PantheonTechnologies/vpptop/stats/api"
	"github.com/PantheonTechnologies/vpptop/stats/local/binapi/dhcp"
	interfaces "github.com/PantheonTechnologies/vpptop/stats/local/binapi/interface"
	"github.com/PantheonTechnologies/vpptop/stats/local/binapi/ip"
	"github.com/PantheonTechnologies/vpptop/stats/local/binapi/vpe"
	"github.com/PantheonTechnologies/vpptop/stats/local/vppcalls"
)

// GetVersion of the local VPP implementation
const VPPVersion = "20.09-rc0"

var localMsgs []govppapi.Message

func init() {
	var msgList []govppapi.Message
	msgList = append(msgList, dhcp.AllMessages()...)
	msgList = append(msgList, interfaces.AllMessages()...)
	msgList = append(msgList, ip.AllMessages()...)
	msgList = append(msgList, vpe.AllMessages()...)
	localMsgs = msgList
}

// HandlerDef is a local handler definition. It is used to validate
// compatibility with the version of the connected VPP
type HandlerDef struct{}

func (d *HandlerDef) IsHandlerCompatible(c *api.VppClient, isRemote bool) (api.HandlerAPI, string, error) {
	ch, err := c.NewAPIChannel()
	if err != nil {
		return nil, "", err
	}
	if err := ch.CheckCompatiblity(localMsgs...); err == nil {
		return NewLocalHandler(c, ch, isRemote), VPPVersion, nil
	}
	return nil, "", nil
}

// Handler makes use of the local implementation to obtain VPP data.
type Handler struct {
	vppCoreCalls      vppcalls.VppCoreAPI
	interfaceVppCalls vppcalls.InterfaceVppAPI
	telemetryVppCalls vppcalls.TelemetryVppAPI
	apiChan           govppapi.Channel
}

// NewLocalHandler returns new instance of the local handler
func NewLocalHandler(c *api.VppClient, ch govppapi.Channel, isRemote bool) *Handler {
	if isRemote {
		for _, msg := range localMsgs {
			gob.Register(msg)
		}
	}
	return &Handler{
		vppCoreCalls:      vppcalls.NewVppCoreHandler(c.Connection(), ch),
		interfaceVppCalls: vppcalls.NewInterfaceHandler(ch),
		telemetryVppCalls: vppcalls.NewTelemetryHandler(c.Connection(), c.Stats()),
		apiChan:           ch,
	}
}

func (h *Handler) DumpInterfaces(ctx context.Context) (map[uint32]*api.InterfaceDetails, error) {
	return h.interfaceVppCalls.DumpInterfaces(ctx)
}

func (h *Handler) DumpInterfaceStats(ctx context.Context) (*govppapi.InterfaceStats, error) {
	return h.telemetryVppCalls.GetInterfaceStats(ctx)
}

func (h *Handler) DumpNodeCounters(ctx context.Context) (*api.NodeCounterInfo, error) {
	return h.telemetryVppCalls.GetNodeCounters(ctx)
}

func (h *Handler) DumpRuntimeInfo(ctx context.Context) (*api.RuntimeInfo, error) {
	return h.telemetryVppCalls.GetRuntimeInfo(ctx)
}

func (h *Handler) RunCli(ctx context.Context, cmd string) (string, error) {
	return h.vppCoreCalls.RunCli(ctx, cmd)
}

func (h *Handler) DumpPlugins(ctx context.Context) ([]api.PluginInfo, error) {
	return h.vppCoreCalls.GetPlugins(ctx)
}

func (h *Handler) DumpVersion(ctx context.Context) (*api.VersionInfo, error) {
	return h.vppCoreCalls.GetVersion(ctx)
}

func (h *Handler) DumpSession(ctx context.Context) (*api.SessionInfo, error) {
	return h.vppCoreCalls.GetSession(ctx)
}
func (h *Handler) DumpThreads(_ context.Context) ([]api.ThreadData, error) {
	req := &vpe.ShowThreads{}
	reply := &vpe.ShowThreadsReply{}
	if err := h.apiChan.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	result := make([]api.ThreadData, len(reply.ThreadData))
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

func (h *Handler) Close() {
	if h.apiChan != nil {
		h.apiChan.Close()
	}
}
