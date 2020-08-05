package stats

import (
	"context"
	"git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging/logrus"
	govppcalls "go.ligato.io/vpp-agent/v2/plugins/govppmux/vppcalls"
	telemetrycalls "go.ligato.io/vpp-agent/v2/plugins/telemetry/vppcalls"
	ifplugincalls "go.ligato.io/vpp-agent/v2/plugins/vpp/ifplugin/vppcalls"
)

// AgentHandler uses Ligato VPP-Agent interface and telemetry handlers
// to obtain data from VPP
type AgentHandler struct {
	vppCoreCalls      govppcalls.VppCoreAPI
	interfaceVppCalls ifplugincalls.InterfaceVppAPI
	telemetryVppCalls telemetrycalls.TelemetryVppAPI
}

// NewCompatibleAgentHandler creates a new instance of the agent handler
func NewCompatibleAgentHandler(c *vppClient) *AgentHandler {
	return &AgentHandler{
		vppCoreCalls:      govppcalls.CompatibleHandler(c),
		interfaceVppCalls: ifplugincalls.CompatibleInterfaceVppHandler(c, logrus.NewLogger("")),
		telemetryVppCalls: telemetrycalls.CompatibleTelemetryHandler(c),
	}
}

func (h *AgentHandler) DumpInterfaces(ctx context.Context) (map[uint32]*ifplugincalls.InterfaceDetails, error) {
	return h.interfaceVppCalls.DumpInterfaces(ctx)
}

func (h *AgentHandler) GetInterfaceStats(ctx context.Context) (*api.InterfaceStats, error) {
	return h.telemetryVppCalls.GetInterfaceStats(ctx)
}

func (h *AgentHandler) GetNodeCounters(ctx context.Context) (*telemetrycalls.NodeCounterInfo, error) {
	return h.telemetryVppCalls.GetNodeCounters(ctx)
}

func (h *AgentHandler) GetRuntimeInfo(ctx context.Context) (*telemetrycalls.RuntimeInfo, error) {
	return h.telemetryVppCalls.GetRuntimeInfo(ctx)
}

func (h *AgentHandler) RunCli(ctx context.Context, cmd string) (string, error) {
	return h.vppCoreCalls.RunCli(ctx, cmd)
}

func (h *AgentHandler) GetPlugins(ctx context.Context) ([]govppcalls.PluginInfo, error) {
	return h.vppCoreCalls.GetPlugins(ctx)
}

func (h *AgentHandler) GetVersion(ctx context.Context) (*govppcalls.VersionInfo, error) {
	return h.vppCoreCalls.GetVersion(ctx)
}

func (h *AgentHandler) GetSession(ctx context.Context) (*govppcalls.SessionInfo, error) {
	return h.vppCoreCalls.GetSession(ctx)
}
