all::

include Makefile.common

BIN ?= apcupsd_exporter

PROMTOOL_VERSION ?= 2.30.0
PROMTOOL_URL     ?= https://github.com/prometheus/prometheus/releases/download/v$(PROMTOOL_VERSION)/prometheus-$(PROMTOOL_VERSION).$(GO_BUILD_PLATFORM).tar.gz
PROMTOOL         ?= $(FIRST_GOPATH)/bin/promtool

DOCKER_IMAGE_NAME       ?= $(BIN)
DOCKER_IMAGE_TAG        ?= latest
MACH                    ?= $(shell uname -m)

ifeq ($(MACH),x86_64)
ARCH := amd64
else
ifeq ($(MACH),aarch64)
ARCH := arm64
endif
endif

STATICCHECK_IGNORE =

PROMU_CONF := .promu.yml
PROMU := $(FIRST_GOPATH)/bin/promu --config $(PROMU_CONF)

.PHONY: build
build: promu $(BIN)
$(BIN): *.go
	$(PROMU) build --prefix=output

.PHONY: fmt
fmt:
	@echo ">> Running fmt"
	gofmt -l -w -s .

.PHONY: crossbuild
crossbuild: promu
	@echo ">> Running crossbuild"
	GOARCH=amd64 $(PROMU) build --prefix=output/amd64
	GOARCH=arm64 $(PROMU) build --prefix=output/arm64

.PHONY: podman-build
podman-build:
	podman build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .

.PHONY: clean
clean:
	@echo ">> Running clean"
	rm -rf $(BIN) output
