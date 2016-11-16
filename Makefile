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

version/version.go: $(gitdir)/HEAD
	go generate ./version

goes-example: | version/version.go
	go build -o $@ ./goes/goes-example

goes-example-amd64.cpio.xz: | version/version.go
	./scripts/mkinitrd ./goes/goes-example

goes-example-arm.cpio.xz: | version/version.go
	env GOARCH=arm ./scripts/mkinitrd ./goes/goes-example

goes-platina-mk1-bmc-arm.cpio.xz: | version/version.go
	env GOARCH=arm ./scripts/mkinitrd ./goes/goes-platina-mk1-bmc

goes-platina-mk1: | version/version.go
	go build -o $@ -tags "uio_pci_dma debug foxy" -gcflags "-N -l" \
		./goes/goes-platina-mk1

go-wip:
	go build -o $@ -tags "uio_pci_dma debug foxy" -gcflags "-N -l" ./wip/y

.PHONY: clean
clean:
	@rm -f go-* goes-*
	@git clean -d -f
