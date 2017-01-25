#!/usr/bin/make
# make VNET_DEBUG=yes to enable vnet debugging checks and flags for gdb.

gitdir := $(shell git rev-parse --git-dir)

gobuild = $(if $(arch),env GOARCH=$(arch) )go build$(if $(tags),\
-tags "$(tags)")$(if $(gcflags),\
-gcflags "$(gcflags)")$(if $(ldflags),\
-ldflags "$(ldflags)")

diag_tag=$(if $(filter yes,$(diag)), diag)
vnet_debug_tag=$(if $(filter yes,$(VNET_DEBUG)), debug)
vnet_gcflags=$(if $(filter yes,$(VNET_DEBUG)),-N -l)

ALL  = goes-example
ALL += goes-example-arm
ALL += goes-platina-mk1-bmc
ifneq (,$(wildcard vnet/devices/ethernet/switch/fe1/*.go))
ALL += goes-platina-mk1
ALL += goes-test
ALL += go-wip
endif

.PHONY: all
all: $(ALL)

copyright/copyright.go: LICENSE PATENTS
	go generate ./copyright

version/version.go: $(gitdir)/HEAD
	go generate ./version

goes-example: | copyright/copyright.go version/version.go
	$(gobuild) -o $@ ./main/goes-example

goes-example-arm: arch=arm
goes-example-arm: tags=netgo
goes-example-arm: ldflags=-d
goes-example-arm: | copyright/copyright.go version/version.go
	$(gobuild) -o $@ ./main/goes-example

goes-platina-mk1-bmc: arch=arm
goes-platina-mk1-bmc: tags=netgo$(diag_tag)
goes-platina-mk1-bmc: ldflags=-d
goes-platina-mk1-bmc: | copyright/copyright.go version/version.go
	$(gobuild) -o $@ ./main/$@

goes-platina-mk1: tags=uio_pci_dma foxy$(vnet_debug_tag)$(diag_tag)
goes-platina-mk1: gcflags=$(vnet_gcflags)
goes-platina-mk1: | copyright/copyright.go version/version.go
	$(gobuild) -o $@ ./main/$@
	$(if $(wildcard fe1a.zip),\
	cat fe1a.zip >> $@ && zip -A $@)

goes-test: | copyright/copyright.go version/version.go
	$(gobuild) -o $@ ./main/goes-test

go-wip: tags=uio_pci_dma foxy$(vnet_debug_)$(diag_tag)
go-wip: gcflags=$(vnet_gcflags)
go-wip:
	$(gobuild) -o $@ ./wip/y
	$(if $(wildcard fe1a.zip),\
	cat fe1a.zip >> $@ && zip -A $@)

fe1a.zip: firmware-fe1a
	./firmware-fe1a

firmware-fe1a:
	go build github.com/platinasystems/$@

.PHONY: clean
clean:
	@rm -f go-* goes-*
	@git clean -d -f
