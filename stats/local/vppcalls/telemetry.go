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
	"regexp"
	"strconv"
	"strings"

	govppapi "git.fd.io/govpp.git/api"
	"go.pantheon.tech/vpptop/stats/api"
	"go.pantheon.tech/vpptop/stats/local/binapi/vpe"
	"github.com/pkg/errors"
)

// TelemetryVppAPI defines telemetry-specific methods
type TelemetryVppAPI interface {
	GetInterfaceStats(context.Context) (*govppapi.InterfaceStats, error)
	GetNodeCounters(context.Context) (*api.NodeCounterInfo, error)
	GetRuntimeInfo(context.Context) (*api.RuntimeInfo, error)
	GetThreads(context.Context) ([]api.ThreadData, error)
}

// TelemetryHandler implements TelemetryVppAPI
type TelemetryHandler struct {
	sp     govppapi.StatsProvider
	vpeRpc vpe.RPCService
}

// NewTelemetryHandler returns a new instance of the TelemetryVppAPI
func NewTelemetryHandler(conn govppapi.Connection, sp govppapi.StatsProvider) TelemetryVppAPI {
	return &TelemetryHandler{
		vpeRpc: vpe.NewServiceClient(conn),
		sp:     sp,
	}
}

// Regular expressions used to parse telemetry output
var (
	// 'show runtime'
	runtimeRe = regexp.MustCompile(`Time ([0-9\.e-]+), ([0-9]+) sec internal node vector rate ([0-9\.e-]+) loops/sec ([0-9\.e-]+)\s+` +
		`vector rates in ([0-9\.e-]+), out ([0-9\.e-]+), drop ([0-9\.e-]+), punt ([0-9\.e-]+)\n` +
		`\s+Name\s+State\s+Calls\s+Vectors\s+Suspends\s+Clocks\s+Vectors/Call\s+` +
		`((?:[\w-:\.]+\s+\w+(?:[ -]\w+)*\s+\d+\s+\d+\s+\d+\s+[0-9\.e-]+\s+[0-9\.e-]+\s+)+)`)
	// 'show runtime' items
	runtimeItemsRe = regexp.MustCompile(`([\w-:.]+)\s+(\w+(?:[ -]\w+)*)\s+(\d+)\s+(\d+)\s+(\d+)\s+([0-9.e-]+)\s+([0-9.e-]+)\s+`)
	// 'show node counters'
	nodeCountersRe    = regexp.MustCompile(`^\s+(\d+)\s+([\w-/]+)\s+(\w+(?:[ -]\w+)*)\s+(\w+)\s+$`)
	nodeCountersReOld = regexp.MustCompile(`^\s+(\d+)\s+([\w-/]+)\s+(.+)$`)
)

func (h *TelemetryHandler) GetInterfaceStats(context.Context) (*govppapi.InterfaceStats, error) {
	ifStats := &govppapi.InterfaceStats{}
	err := h.sp.GetInterfaceStats(ifStats)
	if err != nil {
		return nil, err
	}
	return ifStats, nil
}

func (h *TelemetryHandler) GetNodeCounters(ctx context.Context) (*api.NodeCounterInfo, error) {
	var counters []api.NodeCounter
	data, err := h.vpeRpc.CliInband(ctx, &vpe.CliInband{
		Cmd: "show node counters",
	})
	if err != nil {
		return nil, errors.Wrap(err, "VPP CLI command \"show node counters\" failed")
	}
	for i, line := range strings.Split(data.Reply, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if i == 0 {
			fields := strings.Fields(line)
			if (len(fields) == 3 || len(fields) == 4) && fields[0] == "Count" {
				continue
			}
			return nil, fmt.Errorf("invalid header for `show node counters` received: %q", line)
		}
		if matches := nodeCountersRe.FindStringSubmatch(line); len(matches)-1 == 4 {
			fields := matches[1:]
			counters = append(counters, api.NodeCounter{
				Count:    uint64(strToFloat64(fields[0])),
				Node:     fields[1],
				Reason:   fields[2],
				Severity: fields[3],
			})
		} else if matches := nodeCountersReOld.FindStringSubmatch(line); len(matches)-1 == 3 {
			// fallback to older version
			fields := matches[1:]

			counters = append(counters, api.NodeCounter{
				Count:    uint64(strToFloat64(fields[0])),
				Node:     fields[1],
				Reason:   fields[2],
				Severity: "unknown",
			})
		} else {
			return nil, fmt.Errorf("`show node counters` parsing failed line: %q", line)
		}
	}
	return &api.NodeCounterInfo{
		Counters: counters,
	}, nil
}

func (h *TelemetryHandler) GetRuntimeInfo(ctx context.Context) (*api.RuntimeInfo, error) {
	cliResp, err := h.vpeRpc.CliInband(ctx, &vpe.CliInband{
		Cmd: "show runtime",
	})
	if err != nil {
		return nil, errors.Wrap(err, "VPP CLI command \"show runtime\" failed")
	}
	threadMatches := runtimeRe.FindAllStringSubmatch(cliResp.Reply, -1)
	if len(threadMatches) == 0 && cliResp.Reply != "" {
		return nil, fmt.Errorf("invalid command: %q, thread matches: %d", cliResp.Reply, len(threadMatches))
	}

	var threads []api.RuntimeThread
	for _, matches := range threadMatches {
		fields := matches[1:]
		if len(fields) != 9 {
			return nil, fmt.Errorf("invalid runtime data for thread (len=%v): %q", len(fields), matches[0])
		}
		thread := api.RuntimeThread{
			Time:               strToFloat64(fields[0]),
			AvgVectorsPerNode:  strToFloat64(fields[1]),
			LastMainLoops:      uint64(strToFloat64(fields[2])),
			VectorsPerMainLoop: strToFloat64(fields[3]),
			VectorRatesIn:      strToFloat64(fields[4]),
			VectorRatesOut:     strToFloat64(fields[5]),
			VectorRatesDrop:    strToFloat64(fields[6]),
			VectorRatesPunt:    strToFloat64(fields[7]),
		}

		itemMatches := runtimeItemsRe.FindAllStringSubmatch(fields[8], -1)
		for _, matches := range itemMatches {
			fields := matches[1:]
			if len(fields) != 7 {
				return nil, fmt.Errorf("invalid runtime data for thread item: %q", matches[0])
			}
			thread.Items = append(thread.Items, api.RuntimeItem{
				Name:           fields[0],
				State:          strings.Replace(fields[1], " ", "-", -1),
				Calls:          uint64(strToFloat64(fields[2])),
				Vectors:        uint64(strToFloat64(fields[3])),
				Suspends:       uint64(strToFloat64(fields[4])),
				Clocks:         strToFloat64(fields[5]),
				VectorsPerCall: strToFloat64(fields[6]),
			})
		}

		threads = append(threads, thread)
	}

	return &api.RuntimeInfo{
		Threads: threads,
	}, nil
}

func (h *TelemetryHandler) GetThreads(ctx context.Context) ([]api.ThreadData, error) {
	threads, err := h.vpeRpc.ShowThreads(ctx, new(vpe.ShowThreads))
	if err != nil {
		return nil, fmt.Errorf("show threads error: %v", err)
	}

	result := make([]api.ThreadData, len(threads.ThreadData))
	for i := range threads.ThreadData {
		result[i].ID = threads.ThreadData[i].ID
		result[i].Name = threads.ThreadData[i].Name
		result[i].Type = threads.ThreadData[i].Type
		result[i].PID = threads.ThreadData[i].PID
		result[i].Core = threads.ThreadData[i].Core
		result[i].CPUID = threads.ThreadData[i].CPUID
		result[i].CPUSocket = threads.ThreadData[i].CPUSocket
	}

	return result, nil
}

func strToFloat64(s string) float64 {
	// Replace 'k' (thousands) with 'e3' to make it parsable with strconv
	s = strings.Replace(s, "k", "e3", 1)
	s = strings.Replace(s, "K", "e3", 1)
	s = strings.Replace(s, "m", "e6", 1)
	s = strings.Replace(s, "M", "e6", 1)
	s = strings.Replace(s, "g", "e9", 1)
	s = strings.Replace(s, "G", "e9", 1)

	num, err := strconv.ParseFloat(s, 10)
	if err != nil {
		return 0
	}
	return num
}
