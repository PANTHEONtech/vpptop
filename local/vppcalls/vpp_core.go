package vppcalls

import (
	"context"
	"fmt"
	"git.fd.io/govpp.git/api"
	"github.com/PantheonTechnologies/vpptop/local/binapi/vpe"
	"github.com/prometheus/common/log"
	govppcalls "go.ligato.io/vpp-agent/v2/plugins/govppmux/vppcalls"
	"strings"
)

type VppCoreAPI interface {
	RunCli(ctx context.Context, cmd string) (string, error)
	GetPlugins(context.Context) ([]govppcalls.PluginInfo, error)
	GetVersion(context.Context) (*govppcalls.VersionInfo, error)
	GetSession(context.Context) (*govppcalls.SessionInfo, error)
}

// VppCoreHandler implements VppCoreAPI
type VppCoreHandler struct {
	ch     api.Channel
	vpeRpc vpe.RPCService
}

// NewVppCoreHandler returns a new instance of the VppCoreAPI
func NewVppCoreHandler(ch api.Channel) VppCoreAPI {
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

func (h VppCoreHandler) GetPlugins(ctx context.Context) ([]govppcalls.PluginInfo, error) {
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

	var plugins []govppcalls.PluginInfo
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
		plugin := govppcalls.PluginInfo{
			Name:        strings.TrimSuffix(fields[1], "_plugin.so"),
			Path:        fields[1],
			Version:     fields[2],
			Description: strings.Join(fields[3:], " "),
		}
		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

func (h VppCoreHandler) GetVersion(ctx context.Context) (*govppcalls.VersionInfo, error) {
	version, err := h.vpeRpc.ShowVersion(ctx, new(vpe.ShowVersion))
	if err != nil {
		return nil, err
	}
	info := &govppcalls.VersionInfo{
		Program:        version.Program,
		Version:        version.Version,
		BuildDate:      version.BuildDate,
		BuildDirectory: version.BuildDirectory,
	}
	return info, nil
}

func (h VppCoreHandler) GetSession(ctx context.Context) (*govppcalls.SessionInfo, error) {
	ctrlPing, err := h.vpeRpc.ControlPing(ctx, new(vpe.ControlPing))
	if err != nil {
		return nil, fmt.Errorf("control ping error: %v", err)
	}
	info := &govppcalls.SessionInfo{
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
