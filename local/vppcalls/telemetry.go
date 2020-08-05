package vppcalls

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"git.fd.io/govpp.git/api"
	"github.com/PantheonTechnologies/vpptop/local/binapi/vpe"
	"github.com/pkg/errors"
	telemetrycalls "go.ligato.io/vpp-agent/v2/plugins/telemetry/vppcalls"
	"go.ligato.io/vpp-agent/v2/plugins/vpp"
)

// TelemetryVppAPI defines telemetry-specific methods
type TelemetryVppAPI interface {
	GetInterfaceStats(context.Context) (*api.InterfaceStats, error)
	GetNodeCounters(context.Context) (*telemetrycalls.NodeCounterInfo, error)
	GetRuntimeInfo(context.Context) (*telemetrycalls.RuntimeInfo, error)
}

// TelemetryHandler implements TelemetryVppAPI
type TelemetryHandler struct {
	sp      api.StatsProvider
	ifStats api.InterfaceStats
	vpeRpc  vpe.RPCService
}

// NewTelemetryHandler returns a new instance of the TelemetryVppAPI
func NewTelemetryHandler(ch api.Channel, c vpp.Client) TelemetryVppAPI {
	return &TelemetryHandler{
		vpeRpc: vpe.NewServiceClient(ch),
		sp:     c.Stats(),
	}
}

// Regular expressions used to parse telemetry output
var (
	// 'show runtime'
	runtimeRe = regexp.MustCompile(`(?:-+\n)?(?:Thread (\d+) (\w+)(?: \(lcore \d+\))?\n)?` +
		`Time ([0-9\.e-]+), average vectors/node ([0-9\.e-]+), last (\d+) main loops ([0-9\.e-]+) per node ([0-9\.e-]+)\s+` +
		`vector rates in ([0-9\.e-]+), out ([0-9\.e-]+), drop ([0-9\.e-]+), punt ([0-9\.e-]+)\n` +
		`\s+Name\s+State\s+Calls\s+Vectors\s+Suspends\s+Clocks\s+Vectors/Call\s+(?:Perf Ticks\s+)?` +
		`((?:[\w-:\.]+\s+\w+(?:[ -]\w+)*\s+\d+\s+\d+\s+\d+\s+[0-9\.e-]+\s+[0-9\.e-]+\s+)+)`)
	// 'show runtime' items
	runtimeItemsRe = regexp.MustCompile(`([\w-:.]+)\s+(\w+(?:[ -]\w+)*)\s+(\d+)\s+(\d+)\s+(\d+)\s+([0-9.e-]+)\s+([0-9.e-]+)\s+`)
	// 'show node counters'
	nodeCountersRe = regexp.MustCompile(`^\s+(\d+)\s+([\w-/]+)\s+(.+)$`)
)

func (h *TelemetryHandler) GetInterfaceStats(context.Context) (*api.InterfaceStats, error) {
	err := h.sp.GetInterfaceStats(&h.ifStats)
	if err != nil {
		return nil, err
	}
	return &h.ifStats, nil
}

func (h *TelemetryHandler) GetNodeCounters(ctx context.Context) (*telemetrycalls.NodeCounterInfo, error) {
	var counters []telemetrycalls.NodeCounter
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
			if len(fields) != 3 || fields[0] != "Count" {
				return nil, fmt.Errorf("invalid header for `show node counters` received: %q", line)
			}
			continue
		}
		matches := nodeCountersRe.FindStringSubmatch(line)
		if len(matches)-1 != 3 {
			return nil, fmt.Errorf("`show node counters` parsing failed line: %q", line)
		}
		fields := matches[1:]

		counters = append(counters, telemetrycalls.NodeCounter{
			Value: uint64(strToFloat64(fields[0])),
			Node:  fields[1],
			Name:  fields[2],
		})
	}
	return &telemetrycalls.NodeCounterInfo{
		Counters: counters,
	}, nil
}

func (h *TelemetryHandler) GetRuntimeInfo(ctx context.Context) (*telemetrycalls.RuntimeInfo, error) {
	cliResp, err := h.vpeRpc.CliInband(ctx, &vpe.CliInband{
		Cmd: "show runtime",
	})
	if err != nil {
		return nil, errors.Wrap(err, "VPP CLI command \"show runtime\" failed")
	}
	threadMatches := runtimeRe.FindAllStringSubmatch(cliResp.Reply, -1)
	if len(threadMatches) == 0 && cliResp.Reply != "" {
		return nil, fmt.Errorf("invalid command: %q", cliResp.Reply)
	}

	var threads []telemetrycalls.RuntimeThread
	for _, matches := range threadMatches {
		fields := matches[1:]
		if len(fields) != 12 {
			return nil, fmt.Errorf("invalid runtime data for thread (len=%v): %q", len(fields), matches[0])
		}
		thread := telemetrycalls.RuntimeThread{
			ID:                  uint(strToFloat64(fields[0])),
			Name:                fields[1],
			Time:                strToFloat64(fields[2]),
			AvgVectorsPerNode:   strToFloat64(fields[3]),
			LastMainLoops:       uint64(strToFloat64(fields[4])),
			VectorsPerMainLoop:  strToFloat64(fields[5]),
			VectorLengthPerNode: strToFloat64(fields[6]),
			VectorRatesIn:       strToFloat64(fields[7]),
			VectorRatesOut:      strToFloat64(fields[8]),
			VectorRatesDrop:     strToFloat64(fields[9]),
			VectorRatesPunt:     strToFloat64(fields[10]),
		}

		itemMatches := runtimeItemsRe.FindAllStringSubmatch(fields[11], -1)
		for _, matches := range itemMatches {
			fields := matches[1:]
			if len(fields) != 7 {
				return nil, fmt.Errorf("invalid runtime data for thread item: %q", matches[0])
			}
			thread.Items = append(thread.Items, telemetrycalls.RuntimeItem{
				Name:           fields[0],
				State:          fields[1],
				Calls:          uint64(strToFloat64(fields[2])),
				Vectors:        uint64(strToFloat64(fields[3])),
				Suspends:       uint64(strToFloat64(fields[4])),
				Clocks:         strToFloat64(fields[5]),
				VectorsPerCall: strToFloat64(fields[6]),
			})
		}

		threads = append(threads, thread)
	}

	return &telemetrycalls.RuntimeInfo{
		Threads: threads,
	}, nil
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
