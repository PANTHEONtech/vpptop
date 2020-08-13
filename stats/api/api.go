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

package api

import (
	"context"
	govppapi "git.fd.io/govpp.git/api"
)

// VppProviderAPI uses VPPTop app to manage VPP connection and retrieve
// various VPP data and statistics in proper format. VppProvider retrieves
// data via respective handler
type VppProviderAPI interface {
	// Connect to the VPP either using provided socket, or remotely
	// with help of remote Address
	Connect(soc string) error
	ConnectRemote(rAddr string) error

	// Disconnect from the VPP
	Disconnect()

	// Get various VPP data (version, interfaces, ...)
	GetVersion() string
	GetInterfaces(ctx context.Context) ([]Interface, error)
	GetNodes(ctx context.Context) ([]Node, error)
	GetErrors(ctx context.Context) ([]Error, error)
	GetMemory(ctx context.Context) ([]string, error)
	GetThreads(ctx context.Context) ([]ThreadData, error)

	// Clear VPP counters
	ClearInterfaceCounters(ctx context.Context) error
	ClearRuntimeCounters(ctx context.Context) error
	ClearErrorCounters(ctx context.Context) error
}

// HandlerAPI uses appropriate underlying implementation (either local
// or via Ligato agent) to obtain all stats required to be displayed by the VPPTop.
// It effectively replaces specific plugin-based handlers (interface, telemetry... )
type HandlerAPI interface {
	// RunCli sends CLI command to VPP
	RunCli(ctx context.Context, cmd string) (string, error)

	// DumpInterfaces retrieves VPP interface data and returns them as
	// a northbound interface data
	DumpInterfaces(ctx context.Context) (map[uint32]*InterfaceDetails, error)

	// DumpInterfaceStats retrieves interface statistics
	DumpInterfaceStats(context.Context) (*govppapi.InterfaceStats, error)

	// DumpNodeCounters retrieves information about VPP node counters
	DumpNodeCounters(context.Context) (*NodeCounterInfo, error)

	// DumpRuntimeInfo retrieves node's runtime info
	DumpRuntimeInfo(context.Context) (*RuntimeInfo, error)

	// DumpPlugins retrieves info about loaded VPP plugins.
	DumpPlugins(context.Context) ([]PluginInfo, error)

	// DumpVersion retrieves info about VPP version.
	DumpVersion(context.Context) (*VersionInfo, error)

	// DumpSession retrieves info about active session
	DumpSession(context.Context) (*SessionInfo, error)

	// DumpThreads retrieves info about VPP threads
	DumpThreads(context.Context) ([]ThreadData, error)

	// Close the handler gracefully
	Close()
}

// HandlerDef is a handler definition - it verifies whether the definition is compatible
// with connected VPP version. If so, the flag is set to 'true' and the handler is returned.
// Remote handler in addition also registers VPP API message type records.
type HandlerDef interface {
	IsHandlerCompatible(c *VppClient, isRemote bool) (HandlerAPI, bool, error)
}

type Node = RuntimeItem
type Error = NodeCounter

// InterfaceDetails contains data about retrieved interface
type InterfaceDetails struct {
	Name         string
	InternalName string
	SwIfIndex    uint32
	IsEnabled    bool
	IPAddresses  []string
	MTU          []uint32
}

// Interface contains interface data mandatory for the VPPTop
// including interface counters
type Interface struct {
	govppapi.InterfaceCounters
	IPAddrs []string
	State   string
	MTU     []uint32
}

// VPPInfo basic information about the connected VPP
type VPPInfo struct {
	Connected   bool
	VersionInfo VersionInfo
	SessionInfo SessionInfo
	Plugins     []PluginInfo
}

// VersionInfo is a VPP version
type VersionInfo struct {
	Program        string
	Version        string
	BuildDate      string
	BuildDirectory string
}

// SessionInfo contains data about VPP session
type SessionInfo struct {
	PID       uint32
	ClientIdx uint32
	Uptime    float64
}

// PluginInfo represents a single VPP plugin
type PluginInfo struct {
	Name        string
	Path        string
	Version     string
	Description string
}

// NodeCounterInfo contains telemetry data about VPP node counters
type NodeCounterInfo struct {
	Counters []NodeCounter
}

// NodeCounter (Error) is a single node counter entry type
type NodeCounter struct {
	Value uint64 `json:"value"`
	Node  string `json:"node"`
	Name  string `json:"name"`
}

// RuntimeInfo contains telemetry data about VPP runtime
type RuntimeInfo struct {
	Threads []RuntimeThread
}

// RuntimeThread is a set of thread-related data with Nodes
type RuntimeThread struct {
	ID                  uint          `json:"id"`
	Name                string        `json:"name"`
	Time                float64       `json:"time"`
	AvgVectorsPerNode   float64       `json:"avg_vectors_per_node"`
	LastMainLoops       uint64        `json:"last_main_loops"`
	VectorsPerMainLoop  float64       `json:"vectors_per_main_loop"`
	VectorLengthPerNode float64       `json:"vector_length_per_node"`
	VectorRatesIn       float64       `json:"vector_rates_in"`
	VectorRatesOut      float64       `json:"vector_rates_out"`
	VectorRatesDrop     float64       `json:"vector_rates_drop"`
	VectorRatesPunt     float64       `json:"vector_rates_punt"`
	Items               []RuntimeItem `json:"items"`
}

// RuntimeItem (Node) represents a single node in thread
type RuntimeItem struct {
	Index          uint    `json:"index"`
	Name           string  `json:"name"`
	State          string  `json:"state"`
	Calls          uint64  `json:"calls"`
	Vectors        uint64  `json:"vectors"`
	Suspends       uint64  `json:"suspends"`
	Clocks         float64 `json:"clocks"`
	VectorsPerCall float64 `json:"vectors_per_call"`
}

// ThreadData wraps all thread data counters.
type ThreadData struct {
	ID        uint32
	Name      string
	Type      string
	PID       uint32
	CPUID     uint32
	Core      uint32
	CPUSocket uint32
}
