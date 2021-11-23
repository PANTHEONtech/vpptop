module github.com/PantheonTechnologies/vpptop

go 1.13

require (
	git.fd.io/govpp.git v0.3.6-0.20210223122847-4459b648e9fb
	github.com/gizak/termui/v3 v3.1.0
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.4.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v0.0.7
	go.ligato.io/cn-infra/v2 v2.5.0-alpha.0.20200313154441-b0d4c1b11c73
	go.ligato.io/vpp-agent/v3 v3.2.0-alpha.0.20201029162139-6e02c60eaa76
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	k8s.io/api v0.0.0-20191016225839-816a9b7df678
	k8s.io/apimachinery v0.0.0-20191017185446-6e68a40eebf9
	k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
	k8s.io/utils v0.0.0-20191010214722-8d271d903fe4 // indirect
)

replace github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2 // Vulnerable versions: < 1.0.2
