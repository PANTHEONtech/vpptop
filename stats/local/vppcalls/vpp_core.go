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
	"context"
	"fmt"
	govppapi "git.fd.io/govpp.git/api"
	"github.com/PantheonTechnologies/vpptop/stats/api"
	"github.com/PantheonTechnologies/vpptop/stats/local/binapi/vpe"
	"github.com/prometheus/common/log"
	"strings"
)

// VppCoreAPI defines vpe-specific methods
type VppCoreAPI interface {
	RunCli(ctx context.Context, cmd string) (string, error)
	GetPlugins(context.Context) ([]api.PluginInfo, error)
	GetVersion(context.Context) (*api.VersionInfo, error)
	GetSession(context.Context) (*api.SessionInfo, error)
}

// VppCoreHandler implements VppCoreAPI
type VppCoreHandler struct {
	ch     govppapi.Channel
	vpeRpc vpe.RPCService
}

// NewVppCoreHandler returns a new instance of the VppCoreAPI
func NewVppCoreHandler(ch govppapi.Channel) VppCoreAPI {
	h := &VppCoreHandler{
		vpeRpc: vpe.NewServiceClient(ch),
		ch:     ch,
	}
	return h
}

func (h VppCoreHandler) RunCli(ctx context.Context, cmd string) (string, error) {
	resp, err := h.vpeRpc.CliInband(ctx, &vpe.CliInband{
		Cmd: cmd,
	})
	if err != nil {
		return "", fmt.Errorf("VPP CLI command %s failed: %v", cmd, err)
	}
	return resp.Reply, nil
}

func (h VppCoreHandler) GetPlugins(ctx context.Context) ([]api.PluginInfo, error) {
	const pluginPathPrefix = "Plugin path is:"

	out, err := h.RunCli(ctx, "show plugins")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(out, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("output of 'show plugins' is empty")
	}

	pluginPathLine := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(pluginPathLine, pluginPathPrefix) {
		return nil, fmt.Errorf("unexpected output for 'show plugins'")
	}
	pluginPath := strings.TrimSpace(strings.TrimPrefix(pluginPathLine, pluginPathPrefix))
	if len(pluginPath) == 0 {
		return nil, fmt.Errorf("plugin path not found in output for 'show plugins'")
	}

	var plugins []api.PluginInfo
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		var i int
		if _, err := fmt.Sscanf(fields[0], "%d.", &i); err != nil {
			continue
		}
		if i <= 0 {
			continue
		}
		plugin := api.PluginInfo{
			Name:        strings.TrimSuffix(fields[1], "_plugin.so"),
			Path:        fields[1],
			Version:     fields[2],
			Description: strings.Join(fields[3:], " "),
		}
		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

func (h VppCoreHandler) GetVersion(ctx context.Context) (*api.VersionInfo, error) {
	version, err := h.vpeRpc.ShowVersion(ctx, new(vpe.ShowVersion))
	if err != nil {
		return nil, err
	}
	info := &api.VersionInfo{
		Program:        version.Program,
		Version:        version.Version,
		BuildDate:      version.BuildDate,
		BuildDirectory: version.BuildDirectory,
	}
	return info, nil
}

func (h VppCoreHandler) GetSession(ctx context.Context) (*api.SessionInfo, error) {
	ctrlPing, err := h.vpeRpc.ControlPing(ctx, new(vpe.ControlPing))
	if err != nil {
		return nil, fmt.Errorf("control ping error: %v", err)
	}
	info := &api.SessionInfo{
		PID:       ctrlPing.VpePID,
		ClientIdx: ctrlPing.ClientIndex,
	}

	sysTime, err := h.vpeRpc.ShowVpeSystemTime(ctx, new(vpe.ShowVpeSystemTime))
	if err != nil {
		log.Warn("system time error: %v", err)
	} else {
		info.Uptime = float64(sysTime.VpeSystemTime)
	}
	return info, nil
}
