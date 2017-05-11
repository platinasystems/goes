#!/usr/bin/make
# make V=1 for verbose go builds
# make VNET_DEBUG=yes to enable vnet debugging checks and flags for gdb.

noplugin := yes

gitdir := $(shell git rev-parse --git-dir)

gobuild = $(if $(arch),env GOARCH=$(arch) )go build$(if $(tags),\
-tags "$(tags)")$(if $(gcflags),\
-gcflags "$(gcflags)")$(if $(ldflags),\
-ldflags "$(ldflags)")$(if $(V),\
-v)

fe1_tags=uio_pci_dma foxy
noplugin_tags=$(if $(filter yes,$(noplugin)),$(fe1_tags) noplugin)
diag_tag=$(if $(filter yes,$(diag)), diag)
vnet_debug_tag=$(if $(filter yes,$(VNET_DEBUG)), debug)
vnet_gcflags=$(if $(filter yes,$(VNET_DEBUG)),-N -l)

ALL  = goes-example
ALL += goes-example-arm
ALL += goes-test
ALL += goes-coreboot
ALL += goes-platina-mk1-bmc
ifneq (,$(wildcard vnet/devices/ethernet/switch/fe1/*.go))
ALL += goes-platina-mk1
ALL += go-wip
endif

.PHONY: all
all: $(ALL)

package.go: LICENSE PATENTS
	go generate

../fe1/package.go: ../fe1/LICENSE ../fe1/PATENTS
	go generate ../fe1

../firmware-fe1a/package.go: ../firmware-fe1a/LICENSE
	go generate ../firmware-fe1a

goes-example: | package.go
	$(gobuild) -o $@ ./main/goes-example

goes-example-arm: arch=arm
goes-example-arm: tags=netgo
goes-example-arm: ldflags=-d
goes-example-arm: | package.go
	$(gobuild) -o $@ ./main/goes-example

goes-platina-mk1-bmc: arch=arm
goes-platina-mk1-bmc: tags=netgo$(diag_tag)
goes-platina-mk1-bmc: ldflags=-d
goes-platina-mk1-bmc: | package.go
	$(gobuild) -o $@ ./main/$@

goes-platina-mk1: tags=$(noplugin_tags)$(vnet_debug_tag)$(diag_tag)
goes-platina-mk1: gcflags=$(vnet_gcflags)
goes-platina-mk1: | package.go
	$(gobuild) -o $@ ./main/$@
	$(if $(wildcard fe1a.zip),\
	cat fe1a.zip >> $@ && zip -A $@)

goes-coreboot: | package.go
	$(gobuild) -o $@ ./main/goes-coreboot

goes-test: | package.go
	$(gobuild) -o $@ ./main/goes-test

go-wip: tags=$(fe1_tags) noplugin$(vnet_debug_)$(diag_tag)
go-wip: gcflags=$(vnet_gcflags)
go-wip:
	$(gobuild) -o $@ ./wip/y
	$(if $(wildcard fe1a.zip),\
	cat fe1a.zip >> $@ && zip -A $@)

fe1.so: tags=$(fe1_tags)
fe1.so: | ../fe1/package.go ../firmware-fe1a/package.go
	$(gobuild) -buildmode=plugin ./main/fe1

.PHONY: clean
clean:
	@rm -f go-* goes-*
	@git clean -d -f
