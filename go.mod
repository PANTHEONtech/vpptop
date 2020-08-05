module github.com/PantheonTechnologies/vpptop

go 1.13

require (
	git.fd.io/govpp.git v0.3.5
	github.com/gizak/termui/v3 v3.1.0
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/ligato/cn-infra v2.2.1-0.20191030081411-42c7431fdca1+incompatible
	github.com/lunixbochs/struc v0.0.0-20190916212049-a5c72983bc42
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.4.0
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b // indirect
	github.com/spf13/cobra v0.0.5
	go.ligato.io/vpp-agent/v2 v2.5.0-alpha.0.20200108085843-0e2148d3dd11
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/sys v0.0.0-20200107162124-548cf772de50 // indirect
	k8s.io/api v0.0.0-20191016225839-816a9b7df678
	k8s.io/apimachinery v0.0.0-20191017185446-6e68a40eebf9
	k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
	k8s.io/utils v0.0.0-20191010214722-8d271d903fe4 // indirect
)

replace (
	github.com/ugorji/go v1.1.6 => github.com/ugorji/go v0.0.0-20181204163529-d75b2dcb6bc8
	github.com/ugorji/go/codec => github.com/ugorji/go/codec v0.0.0-20181204163529-d75b2dcb6bc8
)
