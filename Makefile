#!/usr/bin/make
# make V=1 for verbose go builds
# make noplugin=no|yes to force plugin enable/disable
# make diag=yes for enhanced diagnostics
# make VNET_DEBUG=yes to enable vnet debugging checks and flags for gdb.

ALL := goes-example
ALL += goes-example-arm
ALL += goes-test
ALL += goes-coreboot
ALL += goes-platina-mk1-bmc
ALL += goes-platina-mk1

fe1_pkg := github.com/platinasystems/fe1
fe1_dir := $(shell go list -e -f {{.Dir}} $(fe1_pkg))
fe1a_pkg := github.com/platinasystems/firmware-fe1a
fe1a_dir := $(shell go list -e -f {{.Dir}} $(fe1a_pkg))

ifneq (,$(fe1_dir))
  ALL += go-wip
  noplugin = yes
  fe1_packages = $(if $(fe1_dir),$(fe1_dir)/package.go)
  ifneq (,$(fe1a_dir))
    fe1_packages += $(if $(fe1a_dir),$(fe1a_dir)/package.go)
  endif
endif

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

package.go: LICENSE PATENTS
	go generate

ifneq (,$(fe1_dir))
$(fe1_dir)/package.go: $(addprefix $(fe1_dir)/,LICENSE PATENTS)
	go generate $(fe1_pkg)

fe1.so: tags=uio_pci_dma foxy$(VNET_DEBUG_tag)$(diag_tag)
fe1.so: | $(fe1_packages)
	$(gobuild) -buildmode=plugin github.com/platinasystems/go/main/fe1
else
fe1.so: /usr/lib/goes/fe1.so
	cp $< $@
endif

ifneq (,$(fe1a_dir))
$(fe1a_dir)/package.go: $(fe1a_dir)/LICENSE
	go generate $(fe1a_pkg)
endif

goes-example: | package.go
	$(gobuild) -o $@ ./main/goes-example

goes-example-arm: arch=arm
goes-example-arm: tags=netgo
goes-example-arm: ldflags=-d
goes-example-arm: | package.go
	$(gobuild) github.com/platinasystems/go/main/goes-example

goes-platina-mk1-bmc: arch=arm
goes-platina-mk1-bmc: tags=netgo$(diag_tag)
goes-platina-mk1-bmc: ldflags=-d
goes-platina-mk1-bmc: | package.go
	$(gobuild) github.com/platinasystems/go/main/$@

goes-platina-mk1: gcflags=$(if $(VNET_DEBUG_yes),-N -l)

ifeq ($(noplugin),yes)
goes-platina-mk1: tags=uio_pci_dma foxy noplugin$(VNET_DEBUG_tag)$(diag_tag)
goes-platina-mk1: | package.go $(fe1_packages)
	$(gobuild) github.com/platinasystems/go/main/$@
else
goes-platina-mk1: tags=uio_pci_dma foxy$(VNET_DEBUG_tag)$(diag_tag)
goes-platina-mk1: fe1.zip | package.go
	$(gobuild) github.com/platinasystems/go/main/$@
	cat fe1.zip >> $@ && zip -q -A $@
endif

goes-coreboot: | package.go
	$(gobuild) -o $@ ./main/goes-coreboot

goes-test: | package.go
	$(gobuild) -o $@ ./main/goes-test

go-wip: tags=uio_pci_dma foxy noplugin$(VNET_DEBUG_tag)$(diag_tag)
go-wip: gcflags=$(if $(VNET_DEBUG_yes),-N -l)
go-wip:
	$(gobuild) -o $@ ./wip/y

fe1.zip: fe1.so
	zip -q -r $@ fe1.so

.PHONY: clean
clean:
	@rm -f go-* goes-* *.so
	@git clean -d -f
