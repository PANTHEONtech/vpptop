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

package client

import (
	"go.pantheon.tech/vpptop/stats/api"
	"sort"
)

// sortNodeStats sort the slice based specified field
func (app *App) sortNodeStats(nodeStats []api.Node, field int, ascending bool) {
	if field == NoColumn {
		return
	}
	var sortFunc func(i, j int) bool
	switch field {
	case NodeStatNodeName:
		sortFunc = func(i, j int) bool {
			if ascending {
				return nodeStats[i].Name < nodeStats[j].Name
			}
			return nodeStats[i].Name > nodeStats[j].Name

		}
	case NodeStatNodeIndex:
		sortFunc = func(i, j int) bool {
			if ascending {
				return nodeStats[i].Index < nodeStats[j].Index
			}
			return nodeStats[i].Index > nodeStats[j].Index
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
				return nodeStats[i].VectorsPerCall < nodeStats[j].VectorsPerCall
			}
			return nodeStats[i].VectorsPerCall > nodeStats[j].VectorsPerCall
		}
	}
	sort.Slice(nodeStats, sortFunc)
}

// sortInterfaceStats sort the slice based on the specified field
func (app *App) sortInterfaceStats(interfaceStats []api.Interface, field int, ascending bool) {
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
				return interfaceStats[i].MTU[0] < interfaceStats[j].MTU[0]
			}
			return interfaceStats[i].MTU[0] > interfaceStats[j].MTU[0]
		}
	case IfaceStatIfaceMTUIP4:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].MTU[1] < interfaceStats[j].MTU[1]
			}
			return interfaceStats[i].MTU[1] > interfaceStats[j].MTU[1]
		}
	case IfaceStatIfaceMTUIP6:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].MTU[2] < interfaceStats[j].MTU[2]
			}
			return interfaceStats[i].MTU[2] > interfaceStats[j].MTU[2]
		}
	case IfaceStatIfaceMTUMPLS:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].MTU[3] < interfaceStats[j].MTU[3]
			}
			return interfaceStats[i].MTU[3] > interfaceStats[j].MTU[3]
		}
	case IfaceStatIfaceRxPackets:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].Rx.Packets < interfaceStats[j].Rx.Packets
			}
			return interfaceStats[i].Rx.Packets > interfaceStats[j].Rx.Packets
		}
	case IfaceStatIfaceRxBytes:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].Rx.Bytes < interfaceStats[j].Rx.Bytes
			}
			return interfaceStats[i].Rx.Bytes > interfaceStats[j].Rx.Bytes
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
				return interfaceStats[i].RxUnicast.Packets < interfaceStats[j].RxUnicast.Packets
			}
			return interfaceStats[i].RxUnicast.Packets > interfaceStats[j].RxUnicast.Packets
		}
	case IfaceStatIfaceRxUnicastBytes:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].RxUnicast.Bytes < interfaceStats[j].RxUnicast.Bytes
			}
			return interfaceStats[i].RxUnicast.Bytes > interfaceStats[j].RxUnicast.Bytes
		}
	case IfaceStatIfaceRxMulticastPackets:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].RxMulticast.Packets < interfaceStats[j].RxMulticast.Packets
			}
			return interfaceStats[i].RxMulticast.Packets > interfaceStats[j].RxMulticast.Packets
		}
	case IfaceStatIfaceRxMulticastBytes:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].RxMulticast.Bytes < interfaceStats[j].RxMulticast.Bytes
			}
			return interfaceStats[i].RxMulticast.Bytes > interfaceStats[j].RxMulticast.Bytes
		}
	case IfaceStatIfaceRxBroadcastPackets:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].RxBroadcast.Packets < interfaceStats[j].RxBroadcast.Packets
			}
			return interfaceStats[i].RxBroadcast.Packets > interfaceStats[j].RxBroadcast.Packets
		}
	case IfaceStatIfaceRxBroadcastBytes:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].RxBroadcast.Bytes < interfaceStats[j].RxBroadcast.Bytes
			}
			return interfaceStats[i].RxBroadcast.Bytes > interfaceStats[j].RxBroadcast.Bytes
		}
	case IfaceStatIfaceTxPackets:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].Tx.Packets < interfaceStats[j].Tx.Packets
			}
			return interfaceStats[i].Tx.Packets > interfaceStats[j].Tx.Packets
		}
	case IfaceStatIfaceTxBytes:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].Tx.Bytes < interfaceStats[j].Tx.Bytes
			}
			return interfaceStats[i].Tx.Bytes > interfaceStats[j].Tx.Bytes
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
				return interfaceStats[i].TxUnicast.Packets < interfaceStats[j].TxUnicast.Packets
			}
			return interfaceStats[i].TxUnicast.Packets > interfaceStats[j].TxUnicast.Packets
		}
	case IfaceStatIfaceTxUnicastMissBytes:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].TxUnicast.Bytes < interfaceStats[j].TxUnicast.Bytes
			}
			return interfaceStats[i].TxUnicast.Bytes > interfaceStats[j].TxUnicast.Packets
		}
	case IfaceStatIfaceTxMulticastPackets:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].TxMulticast.Packets < interfaceStats[j].TxMulticast.Packets
			}
			return interfaceStats[i].TxMulticast.Packets > interfaceStats[j].TxMulticast.Packets
		}
	case IfaceStatIfaceTxMulticastBytes:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].TxMulticast.Bytes < interfaceStats[j].TxMulticast.Bytes
			}
			return interfaceStats[i].TxMulticast.Bytes > interfaceStats[j].TxMulticast.Bytes
		}
	case IfaceStatIfaceTxBroadcastPackets:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].TxBroadcast.Packets < interfaceStats[j].TxBroadcast.Packets
			}
			return interfaceStats[i].TxBroadcast.Packets > interfaceStats[j].TxBroadcast.Packets
		}
	case IfaceStatIfaceTxBroadcastBytes:
		sortFunc = func(i, j int) bool {
			if ascending {
				return interfaceStats[i].TxBroadcast.Bytes < interfaceStats[j].TxBroadcast.Bytes
			}
			return interfaceStats[i].TxBroadcast.Bytes > interfaceStats[j].TxBroadcast.Bytes
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
func (app *App) sortErrorStats(errorStats []api.Error, field int, ascending bool) {
	if field == NoColumn {
		return
	}
	var sortFunc func(i, j int) bool
	switch field {
	case ErrorStatErrorCounter:
		sortFunc = func(i, j int) bool {
			if ascending {
				return errorStats[i].Count < errorStats[j].Count
			}
			return errorStats[i].Count > errorStats[j].Count
		}
	case ErrorStatErrorNodeName:
		sortFunc = func(i, j int) bool {
			if ascending {
				return errorStats[i].Node < errorStats[j].Node
			}
			return errorStats[i].Node > errorStats[j].Node
		}
	case ErrorStatErrorReason:
		sortFunc = func(i, j int) bool {
			if ascending {
				return errorStats[i].Reason < errorStats[j].Reason
			}
			return errorStats[i].Reason > errorStats[j].Reason
		}
	case ErrorStatErrorSeverity:
		sortFunc = func(i, j int) bool {
			if ascending {
				return errorStats[i].Severity < errorStats[j].Severity
			}
			return errorStats[i].Severity > errorStats[j].Severity
		}
	}
	sort.Slice(errorStats, sortFunc)
}
