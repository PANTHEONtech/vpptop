module go.pantheon.tech/vpptop

go 1.13

require (
	git.fd.io/govpp.git v0.5.0
	github.com/gizak/termui/v3 v3.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.26.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	go.ligato.io/cn-infra/v2 v2.5.0-alpha.0.20220211111933-3d9ff310b1fa
	go.ligato.io/vpp-agent/v3 v3.4.0
	k8s.io/api v0.20.6
	k8s.io/apimachinery v0.20.6
	k8s.io/client-go v0.20.6
)

replace github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2 // Vulnerable versions: < 1.0.2
