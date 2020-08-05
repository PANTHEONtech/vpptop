package stats

import (
	"context"
	"git.fd.io/govpp.git/api"
	"github.com/PantheonTechnologies/vpptop/local"
	"github.com/PantheonTechnologies/vpptop/local/binapi/interfaces"
	"github.com/PantheonTechnologies/vpptop/local/vppcalls"
	govppcalls "go.ligato.io/vpp-agent/v2/plugins/govppmux/vppcalls"
	telemetrycalls "go.ligato.io/vpp-agent/v2/plugins/telemetry/vppcalls"
	"go.ligato.io/vpp-agent/v2/plugins/vpp/binapi"
	ifplugincalls "go.ligato.io/vpp-agent/v2/plugins/vpp/ifplugin/vppcalls"
)

// LocalHandler makes use of the local implementation to obtain VPP data.
type LocalHandler struct {
	vppCoreCalls      vppcalls.VppCoreAPI
	interfaceVppCalls vppcalls.InterfaceVppAPI
	telemetryVppCalls vppcalls.TelemetryVppAPI
}

// NewCompatibleLocalHandler returns new instance of the local handler
func NewCompatibleLocalHandler(c *vppClient) *LocalHandler {
	return &LocalHandler{
		vppCoreCalls:      vppcalls.NewVppCoreHandler(c.apiChan),
		interfaceVppCalls: vppcalls.NewInterfaceHandler(c.apiChan),
		telemetryVppCalls: vppcalls.NewTelemetryHandler(c.apiChan, c),
	}
}

// CheckLocalHandlerCompatible verifies if the local handler is compatible with the
// given VPP version.
func CheckLocalHandlerCompatible(ch api.Channel) (binapi.Version, bool) {
	msgs := interfaces.AllMessages()
	if err := ch.CheckCompatiblity(msgs...); err == nil {
		return local.Version, true
	}
	return "", false
}

func (h *LocalHandler) DumpInterfaces(ctx context.Context) (map[uint32]*ifplugincalls.InterfaceDetails, error) {
	return h.interfaceVppCalls.DumpInterfaces(ctx)
}

func (h *LocalHandler) GetInterfaceStats(ctx context.Context) (*api.InterfaceStats, error) {
	return h.telemetryVppCalls.GetInterfaceStats(ctx)
}

func (h *LocalHandler) GetNodeCounters(ctx context.Context) (*telemetrycalls.NodeCounterInfo, error) {
	return h.telemetryVppCalls.GetNodeCounters(ctx)
}

func (h *LocalHandler) GetRuntimeInfo(ctx context.Context) (*telemetrycalls.RuntimeInfo, error) {
	return h.telemetryVppCalls.GetRuntimeInfo(ctx)
}

func (h *LocalHandler) RunCli(ctx context.Context, cmd string) (string, error) {
	return h.vppCoreCalls.RunCli(ctx, cmd)
}

func (h *LocalHandler) GetPlugins(ctx context.Context) ([]govppcalls.PluginInfo, error) {
	return h.vppCoreCalls.GetPlugins(ctx)
}

func (h *LocalHandler) GetVersion(ctx context.Context) (*govppcalls.VersionInfo, error) {
	return h.vppCoreCalls.GetVersion(ctx)
}

func (h *LocalHandler) GetSession(ctx context.Context) (*govppcalls.SessionInfo, error) {
	return h.vppCoreCalls.GetSession(ctx)
}
