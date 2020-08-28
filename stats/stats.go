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
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"git.fd.io/govpp.git"
	"git.fd.io/govpp.git/adapter"
	"git.fd.io/govpp.git/adapter/statsclient"
	govppapi "git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
	"git.fd.io/govpp.git/proxy"
	"github.com/PantheonTechnologies/vpptop/stats/api"
)

const (
	stateUp   = "up"
	stateDown = "down"
)

// vppProvider provides statistics about VPP such as runtime counters,
// interface counters, error counters and so on
type vppProvider struct {
	// provider clients
	vppClient   *api.VppClient
	statsClient adapter.StatsAPI

	// list of available VPP handler definitions
	handlerDefs []api.HandlerDef
	// interface to the chosen VPP handler
	handler api.HandlerAPI

	vppVersion        *api.VersionInfo
	lastErrorCounters map[string]uint64
}

// NewVppProvider constructs new VppProviderAPI object with available
// VPP version definitions
func NewVppProvider(defs []api.HandlerDef) api.VppProviderAPI {
	return &vppProvider{handlerDefs: defs}
}

// Connect establishes a VPP connection using GoVPP API
func (p *vppProvider) Connect(soc string) error {
	p.lastErrorCounters = make(map[string]uint64)

	p.statsClient = statsclient.NewStatsClient(soc)
	statsConn, err := core.ConnectStats(p.statsClient)
	if err != nil {
		return fmt.Errorf("connection to stats api failed: %v", err)
	}

	vppConn, err := govpp.Connect("")
	if err != nil {
		return fmt.Errorf("connection to govpp failed: %v", err)
	}

	p.vppClient = api.NewVppClient(vppConn, statsConn)

	var handlerFound bool
	for _, handlerDef := range p.handlerDefs {
		handler, isCompatible, err := handlerDef.IsHandlerCompatible(p.vppClient, false)
		if err != nil {
			return err
		}
		if isCompatible {
			p.handler = handler
			handlerFound = true
			break
		}
	}
	if !handlerFound {
		return fmt.Errorf("no compatible handler was found")
	}

	ctx := context.Background()
	plugins, err := p.handler.DumpPlugins(ctx)
	if err != nil {
		return err
	}

	session, err := p.handler.DumpSession(ctx)
	if err != nil {
		return err
	}

	p.vppVersion, err = p.handler.DumpVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get vpp version: %v", err)
	}

	p.vppClient.SetInfo(api.VPPInfo{
		Connected:   true,
		VersionInfo: *p.vppVersion,
		SessionInfo: *session,
		Plugins:     plugins,
	})

	return nil
}

// ConnectRemote connects VPPTop to a remote proxy providing vpp statistics
func (p *vppProvider) ConnectRemote(rAddr string) error {
	p.lastErrorCounters = make(map[string]uint64)

	var err error
	var client *proxy.Client
	for i := 0; i < 3; i++ {
		client, err = proxy.Connect(rAddr)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to raddr %v, reason: %v", rAddr, err)
	}

	statsConn, err := client.NewStatsClient()
	if err != nil {
		return err
	}

	p.vppClient = api.NewProxyClient(client, statsConn)

	var handlerFound bool
	for _, handlerDef := range p.handlerDefs {
		handler, isCompatible, err := handlerDef.IsHandlerCompatible(p.vppClient, true)
		if err != nil {
			return err
		}
		if isCompatible {
			p.handler = handler
			handlerFound = true
			break
		}
	}
	if !handlerFound {
		return fmt.Errorf("no compatible handler was found")
	}

	ctx := context.Background()

	plugins, err := p.handler.DumpPlugins(ctx)
	if err != nil {
		return err
	}

	session, err := p.handler.DumpSession(ctx)
	if err != nil {
		return err
	}

	p.vppVersion, err = p.handler.DumpVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get vpp version: %v", err)
	}

	p.vppClient.SetInfo(api.VPPInfo{
		Connected:   true,
		VersionInfo: *p.vppVersion,
		SessionInfo: *session,
		Plugins:     plugins,
	})

	return nil
}

// Disconnect should be called after Connect, if the connection is no longer needed.
func (p *vppProvider) Disconnect() {
	p.handler.Close()
	if p.vppClient != nil {
		p.vppClient.Disconnect()
		p.vppClient.Close()
	}

	if p.statsClient != nil {
		if err := p.statsClient.Disconnect(); err != nil {
			log.Printf("error disconnecting VPP provider: %v", err)
		}
	}
}

// GetVersion returns the current vpp version.
func (p *vppProvider) GetVersion() string {
	return "VPP version: " + p.vppVersion.Version + "\n" + p.vppVersion.BuildDate
}

// GetNodes returns per node statistics.
func (p *vppProvider) GetNodes(ctx context.Context) ([]api.Node, error) {
	runtimeInfo, err := p.handler.DumpRuntimeInfo(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	threads := runtimeInfo.Threads
	if len(threads) == 0 {
		return nil, errors.New("no runtime counters")
	}

	result := make([]api.Node, 0, len(threads[0].Items))
	for _, thread := range threads {
		for _, item := range thread.Items {
			result = append(result, item)
		}
	}

	return result, nil
}

// GetInterfaces returns per interface statistics.
func (p *vppProvider) GetInterfaces(ctx context.Context) ([]api.Interface, error) {
	var ifStats *govppapi.InterfaceStats
	var ifDetails map[uint32]*api.InterfaceDetails

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
		ifDetails, err = p.handler.DumpInterfaces(ctx)
		errChan <- err
	}()

	go func() {
		defer wg.Done()
		var err error
		ifStats, err = p.handler.DumpInterfaceStats(ctx)
		errChan <- err
	}()

	for err := range errChan {
		if err != nil {
			return nil, fmt.Errorf("request failed: %v", err)
		}
	}

	result := make([]api.Interface, 0, len(ifDetails))
	for _, iface := range ifStats.Interfaces {
		details, ok := ifDetails[iface.InterfaceIndex]
		if !ok {
			continue
		}
		state := stateDown
		if details.IsEnabled {
			state = stateUp
		}
		result = append(result, api.Interface{
			InterfaceCounters: iface,
			IPAddresses:       details.IPAddresses,
			State:             state,
			MTU:               details.MTU,
		})
	}
	return result, nil
}

// GetErrors returns per error statistics.
func (p *vppProvider) GetErrors(ctx context.Context) ([]api.Error, error) {
	nodeCounters, err := p.handler.DumpNodeCounters(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]api.Error, 0)
	for _, counter := range nodeCounters.Counters {
		counter.Value -= p.lastErrorCounters[counter.Node+counter.Name]
		if counter.Value == 0 {
			continue
		}
		result = append(result, counter)
	}

	return result, nil
}

// GetMemory returns memory usage per thread.
func (p *vppProvider) GetMemory(ctx context.Context) ([]string, error) {
	mem, err := p.handler.RunCli(ctx, "show memory main-heap verbose")
	if err != nil {
		return nil, err
	}

	rows := make([]string, 0, 1) // there's going to be at least one thread
	for _, r := range strings.Split(mem, "\n") {
		if r == "" {
			continue
		}

		rows = append(rows, strings.Trim(r, " \n"))
	}

	return rows, nil
}

// GetThreads returns thread data per thread.
func (p *vppProvider) GetThreads(ctx context.Context) ([]api.ThreadData, error) {
	return p.handler.DumpThreads(ctx)
}

// ClearInterfaceCounters resets the counters for the interface.
func (p *vppProvider) ClearInterfaceCounters(ctx context.Context) error {
	if _, err := p.handler.RunCli(ctx, "clear interfaces"); err != nil {
		return fmt.Errorf("request failed: %v", err)
	}

	return nil
}

// ClearRuntimeCounters clears the runtime counters for nodes.
func (p *vppProvider) ClearRuntimeCounters(ctx context.Context) error {
	if _, err := p.handler.RunCli(ctx, "clear runtime"); err != nil {
		return fmt.Errorf("request failed: %v", err)
	}

	return nil
}

// ClearErrorCounters clears the counters for errors.
func (p *vppProvider) ClearErrorCounters(ctx context.Context) error {
	p.updateLastErrors(ctx)
	if _, err := p.handler.RunCli(ctx, "clear errors"); err != nil {
		return fmt.Errorf("request failed: %v", err)
	}

	return nil
}

// updateLastErrors clears the error counters.
func (p *vppProvider) updateLastErrors(ctx context.Context) {
	nodeCounters, err := p.handler.DumpNodeCounters(ctx)
	if err != nil {
		return
	}

	for _, counter := range nodeCounters.Counters {
		if counter.Value == 0 {
			continue
		}
		p.lastErrorCounters[counter.Node+counter.Name] = counter.Value
	}
}
