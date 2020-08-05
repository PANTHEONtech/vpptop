package binapi

//go:generate binapi-generator --output-dir=. --input-file=/usr/share/vpp/api/core/interface.api.json
//go:generate binapi-generator --output-dir=. --input-file=/usr/share/vpp/api/core/ip.api.json
//go:generate binapi-generator --output-dir=. --input-file=/usr/share/vpp/api/core/vpe.api.json
//go:generate binapi-generator --output-dir=. --input-file=/usr/share/vpp/api/plugins/dhcp.api.json

