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

package main

import (
	"sort"

	"github.com/PantheonTechnologies/vpptop/stats"
)

// sortNodeStats sort the slice based specified field
func (app *App) sortNodeStats(nodeStats []stats.Node, field int, ascending bool) {
	if field == NoColumn {
		return
	}
	var sortFunc func(i, j int) bool
	switch field {
	case NodeStatNodeName:
		sortFunc = func(i, j int) bool {
			if ascending {
				return nodeStats[i].NodeName < nodeStats[j].NodeName
			}
			return nodeStats[i].NodeName > nodeStats[j].NodeName

		}
	case NodeStatNodeIndex:
		sortFunc = func(i, j int) bool {
			if ascending {
				return nodeStats[i].NodeIndex < nodeStats[j].NodeIndex
			}
			return nodeStats[i].NodeIndex > nodeStats[j].NodeIndex
		}
	case NodeStatNodeClocks:
		sortFunc = func(i, j int) bool {
			if ascending {
				return nodeStats[i].Clocks < nodeStats[j].Clocks
			}
			return nodeStats[i].Clocks > nodeStats[j].Clocks
		}
	case NodeStatNodeVectors:
		sortFunc = func(i, j int) bool {
			if ascending {
				return nodeStats[i].Vectors < nodeStats[j].Vectors
			}
			return nodeStats[i].Vectors > nodeStats[j].Vectors
		}
	case NodeStatNodeCalls:
		sortFunc = func(i, j int) bool {
			if ascending {
				return nodeStats[i].Calls < nodeStats[j].Calls
			}
			return nodeStats[i].Calls > nodeStats[j].Calls
		}
	case NodeStatNodeSuspends:
		sortFunc = func(i, j int) bool {
			if ascending {
				return nodeStats[i].Suspends < nodeStats[j].Suspends
			}
			return nodeStats[i].Suspends > nodeStats[j].Suspends
		}
	case NodeStatNodeVC:
		sortFunc = func(i, j int) bool {
			if ascending {
				return nodeStats[i].VC < nodeStats[j].VC
			}
			return nodeStats[i].VC > nodeStats[j].VC
		}
	}
	sort.Slice(nodeStats, sortFunc)
}

// sortInterfaceStats sort the slice based on the specified field
func (app *App) sortInterfaceStats(interfaceStats []stats.Interface, field int, ascending bool) {
	if field == NoColumn {
		return
	}
	var sortFunc func(i, j int) bool
	switch field {
	case IfaceStatIfaceName:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].InterfaceName < interfaceStats[j].InterfaceName
			}
			return interfaceStats[i].InterfaceName > interfaceStats[j].InterfaceName
		}
	case IfaceStatIfaceIdx:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].InterfaceIndex < interfaceStats[j].InterfaceIndex
			}
			return interfaceStats[i].InterfaceIndex > interfaceStats[j].InterfaceIndex
		}
	case IfaceStatIfaceState:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].State < interfaceStats[j].State
			}
			return interfaceStats[i].State > interfaceStats[j].State
		}
	case IfaceStatIfaceMTUL3:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].Mtu[0] < interfaceStats[j].Mtu[0]
			}
			return interfaceStats[i].Mtu[0] > interfaceStats[j].Mtu[0]
		}
	case IfaceStatIfaceMTUIP4:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].Mtu[1] < interfaceStats[j].Mtu[1]
			}
			return interfaceStats[i].Mtu[1] > interfaceStats[j].Mtu[1]
		}
	case IfaceStatIfaceMTUIP6:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].Mtu[2] < interfaceStats[j].Mtu[2]
			}
			return interfaceStats[i].Mtu[2] > interfaceStats[j].Mtu[2]
		}
	case IfaceStatIfaceMTUMPLS:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].Mtu[3] < interfaceStats[j].Mtu[3]
			}
			return interfaceStats[i].Mtu[3] > interfaceStats[j].Mtu[3]
		}
	case IfaceStatIfaceRxPackets:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].RxPackets < interfaceStats[j].RxPackets
			}
			return interfaceStats[i].RxPackets > interfaceStats[j].RxPackets
		}
	case IfaceStatIfaceRxBytes:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].RxBytes < interfaceStats[j].RxBytes
			}
			return interfaceStats[i].RxBytes > interfaceStats[j].RxBytes
		}
	case IfaceStatIfaceRxErrors:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].RxErrors < interfaceStats[j].RxErrors
			}
			return interfaceStats[i].RxErrors > interfaceStats[j].RxErrors
		}
	case IfaceStatIfaceRxUnicastPackets:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].RxUnicast[0] < interfaceStats[j].RxUnicast[0]
			}
			return interfaceStats[i].RxUnicast[0] > interfaceStats[j].RxUnicast[0]
		}
	case IfaceStatIfaceRxUnicastBytes:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].RxUnicast[1] < interfaceStats[j].RxUnicast[1]
			}
			return interfaceStats[i].RxUnicast[1] > interfaceStats[j].RxUnicast[1]
		}
	case IfaceStatIfaceRxMulticastPackets:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].RxMulticast[0] < interfaceStats[j].RxMulticast[0]
			}
			return interfaceStats[i].RxMulticast[0] > interfaceStats[j].RxMulticast[0]
		}
	case IfaceStatIfaceRxMulticastBytes:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].RxMulticast[1] < interfaceStats[j].RxMulticast[1]
			}
			return interfaceStats[i].RxMulticast[1] > interfaceStats[j].RxMulticast[1]
		}
	case IfaceStatIfaceRxBroadcastPackets:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].RxBroadcast[0] < interfaceStats[j].RxBroadcast[0]
			}
			return interfaceStats[i].RxBroadcast[0] > interfaceStats[j].RxBroadcast[0]
		}
	case IfaceStatIfaceRxBroadcastBytes:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].RxBroadcast[1] < interfaceStats[j].RxBroadcast[1]
			}
			return interfaceStats[i].RxBroadcast[1] > interfaceStats[j].RxBroadcast[1]
		}
	case IfaceStatIfaceTxPackets:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].TxPackets < interfaceStats[j].TxPackets
			}
			return interfaceStats[i].TxPackets > interfaceStats[j].TxPackets
		}
	case IfaceStatIfaceTxBytes:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].TxBytes < interfaceStats[j].TxBytes
			}
			return interfaceStats[i].TxBytes > interfaceStats[j].TxBytes
		}
	case IfaceStatIfaceTxErrors:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].TxErrors < interfaceStats[j].TxErrors
			}
			return interfaceStats[i].TxErrors > interfaceStats[j].TxErrors
		}
	case IfaceStatIfaceTxUnicastMissPackets:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].TxUnicastMiss[0] < interfaceStats[j].TxUnicastMiss[0]
			}
			return interfaceStats[i].TxUnicastMiss[0] > interfaceStats[j].TxUnicastMiss[0]
		}
	case IfaceStatIfaceTxUnicastMissBytes:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].TxUnicastMiss[1] < interfaceStats[j].TxUnicastMiss[1]
			}
			return interfaceStats[i].TxUnicastMiss[1] > interfaceStats[j].TxUnicastMiss[1]
		}
	case IfaceStatIfaceTxMulticastPackets:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].TxMulticast[0] < interfaceStats[j].TxMulticast[0]
			}
			return interfaceStats[i].TxMulticast[0] > interfaceStats[j].TxMulticast[0]
		}
	case IfaceStatIfaceTxMulticastBytes:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].TxMulticast[1] < interfaceStats[j].TxMulticast[1]
			}
			return interfaceStats[i].TxMulticast[1] > interfaceStats[j].TxMulticast[1]
		}
	case IfaceStatIfaceTxBroadcastPackets:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].TxBroadcast[0] < interfaceStats[j].TxBroadcast[0]
			}
			return interfaceStats[i].TxBroadcast[0] > interfaceStats[j].TxBroadcast[0]
		}
	case IfaceStatIfaceTxBroadcastBytes:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].TxBroadcast[1] < interfaceStats[j].TxBroadcast[1]
			}
			return interfaceStats[i].TxBroadcast[1] > interfaceStats[j].TxBroadcast[1]
		}
	case IfaceStatIfaceDrops:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].Drops < interfaceStats[j].Drops
			}
			return interfaceStats[i].Drops > interfaceStats[j].Drops
		}
	case IfaceStatIfacePunts:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].Punts < interfaceStats[j].Punts
			}
			return interfaceStats[i].Punts > interfaceStats[j].Punts
		}
	case IfaceStatIfaceIP4:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].IP4 < interfaceStats[j].IP4
			}
			return interfaceStats[i].IP4 > interfaceStats[j].IP4
		}
	case IfaceStatIfaceIP6:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].IP6 < interfaceStats[j].IP6
			}
			return interfaceStats[i].IP6 > interfaceStats[j].IP6
		}
	}
	sort.Slice(interfaceStats, sortFunc)
}

// sortErrorStats sorts the slice based on the specified field
func (app *App) sortErrorStats(errorStats []stats.Error, field int, ascending bool) {
	if field == NoColumn {
		return
	}
	var sortFunc func(i, j int) bool
	switch field {
	case ErrorStatErrorCounter:
		sortFunc = func(i, j int) bool {
			if ascending {
				return errorStats[i].Value < errorStats[j].Value
			}
			return errorStats[i].Value > errorStats[j].Value
		}
	case ErrorStatErrorNodeName:
		sortFunc = func(i, j int) bool {
			if ascending {
				return errorStats[i].NodeName < errorStats[j].NodeName
			}
			return errorStats[i].NodeName > errorStats[j].NodeName
		}
	case ErrorStatErrorReason:
		sortFunc = func(i, j int) bool {
			if ascending {
				return errorStats[i].Reason < errorStats[j].Reason
			}
			return errorStats[i].Reason > errorStats[j].Reason
		}
	}
	sort.Slice(errorStats, sortFunc)
}
