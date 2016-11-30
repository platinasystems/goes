#!/usr/bin/make
# make VNET_DEBUG=yes to enable vnet debugging checks and flags for gdb.

gitdir := $(shell git rev-parse --git-dir)

gobuild = $(if $(arch),env GOARCH=$(arch) )go build$(if $(tags),\
-tags "$(tags)")$(if $(gcflags),\
-gcflags "$(gcflags)")$(if $(ldflags),\
-ldflags "$(ldflags)")

VNET_TAGS = uio_pci_dma foxy$(if $(filter yes,$(VNET_DEBUG)), debug)
VNET_GCFLAGS = $(if $(filter yes,$(VNET_DEBUG)),-N -l)

ALL  = goes-example
ALL += goes-example-arm
ALL += goes-platina-mk1-bmc
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
	$(gobuild) -o $@ ./goes/goes-example

goes-example-arm: arch=arm
goes-example-arm: tags=netgo
goes-example-arm: ldflags=-d
goes-example-arm: | copyright/copyright.go version/version.go
	$(gobuild) -o $@ ./goes/goes-example

goes-platina-mk1-bmc: arch=arm
goes-platina-mk1-bmc: tags=netgo
goes-platina-mk1-bmc: ldflags=-d
goes-platina-mk1-bmc: | copyright/copyright.go version/version.go
	$(gobuild) -o $@ ./goes/goes-platina-mk1-bmc

goes-platina-mk1: tags=$(VNET_TAGS)
goes-platina-mk1: gcflags=$(VNET_GCFLAGS)
goes-platina-mk1: | copyright/copyright.go version/version.go
	$(gobuild) -o $@ ./goes/goes-platina-mk1

go-wip: tags=$(VNET_TAGS)
go-wip: gcflags=$(VNET_GCFLAGS)
go-wip:
	$(gobuild) -o $@ ./wip/y

.PHONY: clean
clean:
	@rm -f go-* goes-*
	@git clean -d -f
