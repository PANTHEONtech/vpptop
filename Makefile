PROJECT	:= VPPTop
VERSION	?= $(shell git describe --tags)
COMMIT	?= $(shell git rev-parse HEAD)
BUILD_DATE	?= $(shell date +%s)
VPP_API_DIR	?= /usr/share/vpp/api

LDFLAGS = -w -s \
	-X $(GOPKG)/pkg/version.app=$(PROJECT) \
	-X $(GOPKG)/pkg/version.version=$(VERSION) \
	-X $(GOPKG)/pkg/version.gitCommit=$(COMMIT) \
	-X $(GOPKG)/pkg/version.buildDate=$(BUILD_DATE)

build: ## Build VPPTop binaries
	@echo "# building ${PROJECT} ${VERSION}"
	go build -v -ldflags "${LDFLAGS}"

install: ## Install VPPTop binaries
	@echo "# building ${PROJECT} ${VERSION}"
	go install -ldflags "${LDFLAGS}"

generate-binapi:
	@echo "# generating binapi using input from ${VPP_API_DIR}"
	@./scripts/binapigen.sh ${VPP_API_DIR}