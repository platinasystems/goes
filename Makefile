#!/usr/bin/make
# make V=1 for verbose go builds
# make noplugin=no|yes to force plugin enable/disable
# make diag=yes for enhanced diagnostics
# make VNET_DEBUG=yes to enable vnet debugging checks and flags for gdb.

fe1_pkg  := github.com/platinasystems/fe1
fe1_dir  := $(shell go list -e -f {{.Dir}} $(fe1_pkg))
fe1_gen  := $(if $(fe1_dir),$(fe1_dir)/package.go)
fe1_tags := uio_pci_dma foxy
libfe1so := $(wildcard /usr/lib/goes/fe1.so)

fe1a_pkg := github.com/platinasystems/firmware-fe1a
fe1a_dir := $(shell go list -e -f {{.Dir}} $(fe1a_pkg))
fe1a_gen := $(if $(fe1a_dir),$(fe1a_dir)/package.go)

ALL := goes-example
ALL += goes-example-arm
ALL += goes-test
ALL += goes-coreboot
ALL += goes-platina-mk1-bmc
ALL += goes-platina-mk1-installer
ALL += $(if $(fe1_dir),go-wip)

noplugin := $(if $(fe1_dir),yes,no)
noplugin_yes:=$(filter yes,$(noplugin))
noplugin_tag:=$(if $(noplugin_yes), noplugin)

diag_yes:=$(filter yes,$(diag))
diag_tag:=$(if $(diag_yes), diag)

VNET_DEBUG_yes:=$(filter yes,$(VNET_DEBUG))
VNET_DEBUG_tag:=$(if $(VNET_DEBUG_yes), debug)

gobuild = $(if $(arch),env GOARCH=$(arch) )go build$(if $(tags),\
-tags "$(tags)")$(if $(gcflags),\
-gcflags "$(gcflags)")$(if $(ldflags),\
-ldflags "$(ldflags)")$(if $(V),\
-v)

.PHONY: all
all: $(ALL)

goes-example: | package.go
	$(gobuild) ./main/$@

goes-example-arm: arch=arm
goes-example-arm: tags=netgo
goes-example-arm: ldflags=-d
goes-example-arm: | package.go
	$(gobuild) -o $@ ./main/goes-example

goes-platina-mk1-bmc: arch=arm
goes-platina-mk1-bmc: tags=netgo$(diag_tag)
goes-platina-mk1-bmc: ldflags=-d
goes-platina-mk1-bmc: | package.go
	$(gobuild) ./main/$@

goes-platina-mk1-installer: goes-platina-mk1.zip
	$(gobuild) -ldflags -d -o $@ ./main/goes-installer
	cat $< >> $@
	zip -q -A $@

goes-platina-mk1: gcflags=$(if $(VNET_DEBUG_yes),-N -l)
goes-platina-mk1: tags=$(fe1_tags)$(noplugin_tag)$(VNET_DEBUG_tag)$(diag_tag)
goes-platina-mk1: | $(if $(noplugin_yes),$(fe1_gen) $(fe1a_gen)) package.go
	$(gobuild) ./main/$@

goes-platina-mk1.zip: $(if $(noplugin_yes),,fe1.so) goes-platina-mk1
	@rm -f $@
	zip -q $@ $^

goes-coreboot: | package.go
	$(gobuild) ./main/goes-coreboot

goes-test: | package.go
	$(gobuild) ./main/goes-test

go-wip: tags=$(fe1_tags)$(noplugin_tag)$(VNET_DEBUG_tag)$(diag_tag)
go-wip: gcflags=$(if $(VNET_DEBUG_yes),-N -l)
go-wip:
	$(gobuild) -o $@ ./wip/y

package.go: LICENSE PATENTS
	go generate

fe1.so: tags=$(fe1_tags)$(VNET_DEBUG_tag)$(diag_tag)
fe1.so: $(if $(fe1_dir),| $(fe1_gen) $(fe1a_gen),$(libfe1so))
	$(if $(fe1_dir),$(gobuild) -buildmode=plugin ./main/fe1,\
		$(if $(libfe1so),cp $(libfe1so),touch) $@)

$(fe1_gen): $(addprefix $(fe1_dir)/,LICENSE PATENTS)
	go generate $(fe1_pkg)

$(fe1a_gen): $(fe1a_dir)/LICENSE
	go generate $(fe1a_pkg)

.PHONY: clean
clean:
	@rm -f go-* goes-* *.so *.zip
	@git clean -d -f
