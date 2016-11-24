#!/usr/bin/make

gitdir := $(shell git rev-parse --git-dir)

ALL  = goes-example
ALL += goes-example-arm.cpio.xz
ALL += goes-platina-mk1-bmc-arm.cpio.xz
ifneq (,$(wildcard vnet/devices/ethernet/switch/fe1/*.go))
ALL += goes-platina-mk1
ALL += go-wip
endif

.PHONY: all
all: $(ALL)

copyright/copyright.go: LICENSE PATENTS
	go generate ./copyright

version/version.go: $(gitdir)/HEAD
	go generate ./version

goes-example: | copyright/copyright.go version/version.go
	go build -o $@ ./goes/goes-example

goes-example-amd64.cpio.xz: | copyright/copyright.go version/version.go
	./scripts/mkinitrd ./goes/goes-example

goes-example-arm.cpio.xz: | copyright/copyright.go version/version.go
	env GOARCH=arm ./scripts/mkinitrd ./goes/goes-example

goes-platina-mk1-bmc-arm.cpio.xz: | copyright/copyright.go version/version.go
	env GOARCH=arm ./scripts/mkinitrd ./goes/goes-platina-mk1-bmc

# make VNET_DEBUG=yes to enable vnet debugging checks and flags for gdb.
VNET_DEBUG = no

VNET_TAGS_yes = debug
VNET_GO_BUILD_FLAGS = -tags "uio_pci_dma foxy $(VNET_TAGS_$(VNET_DEBUG))"

VNET_GO_BUILD_FLAGS_yes = -gcflags "-N -l"
VNET_GO_BUILD_FLAGS += $(VNET_GO_BUILD_FLAGS_$(VNET_DEBUG))

goes-platina-mk1: | copyright/copyright.go version/version.go
	go build -o $@ $(VNET_GO_BUILD_FLAGS) ./goes/goes-platina-mk1

go-wip:
	go build -o $@ $(VNET_GO_BUILD_FLAGS) ./wip/y

.PHONY: clean
clean:
	@rm -f go-* goes-*
	@git clean -d -f
