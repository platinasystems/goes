#!/usr/bin/make

ALL = goes-platina-mk1 goes-platina-mk1-installer goes-platina-mk1.test go-wip

reltopdir := $(shell git rev-parse --show-cdup)

include $(reltopdir)go.mk
include $(reltopdir)fe1.mk

goes-platina-mk1-installer: tags=netgo
goes-platina-mk1-installer: goes-platina-mk1.zip
	$(gobuild) -ldflags -d -o $@ ../goes-installer
	cat $< >> $@
	zip -q -A $@

goes-platina-mk1: gcflags=$(if $(VNET_DEBUG_yes),-N -l)
goes-platina-mk1: tags=vfio$(plugin_tag)$(VNET_DEBUG_tag)$(diag_tag)
goes-platina-mk1:
	$(gobuild)

goes-platina-mk1.test: gcflags=$(if $(VNET_DEBUG_yes),-N -l)
goes-platina-mk1.test: tags=vfio$(plugin_tag)$(VNET_DEBUG_tag)$(diag_tag)
goes-platina-mk1.test:
	$(gotest) -c

goes-platina-mk1.zip: $(plugin_fe1) goes-platina-mk1
	@rm -f $@
	zip -q $@ $^

go-wip: tags=vfio foxy$(noplugin_tag)$(VNET_DEBUG_tag)$(diag_tag)
go-wip: gcflags=$(if $(VNET_DEBUG_yes),-N -l)
go-wip:
	$(gobuild) -o $@ $(reltopdir)wip/y
