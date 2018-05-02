#!/usr/bin/make

ALL = goes-platina-mk1 goes-platina-mk1-installer goes-platina-mk1.test

reltopdir := $(shell git rev-parse --show-cdup)

include $(reltopdir)go.mk
include $(reltopdir)fe1.mk

vnet_tags=$(novfio_tag)$(plugin_tag)$(VNET_DEBUG_tag)$(diag_tag)

goes-platina-mk1-installer: tags=netgo
goes-platina-mk1-installer: goes-platina-mk1.zip
	$(gobuild) -ldflags -d -o $@ ../goes-installer
	cat $< >> $@
	zip -q -A $@

goes-platina-mk1: gcflags=$(if $(VNET_DEBUG_yes),-N -l)
goes-platina-mk1: tags=$(strip $(vnet_tags))
goes-platina-mk1:
	$(gobuild)

goes-platina-mk1.test: gcflags=$(if $(VNET_DEBUG_yes),-N -l)
goes-platina-mk1.test: tags=$(strip $(vnet_tags))
goes-platina-mk1.test:
	$(gotest) -c

goes-platina-mk1.zip: $(plugin_fe1) goes-platina-mk1
	@rm -f $@
	zip -q $@ $^

go-wip: tags=foxy$(vnet_tags)
go-wip: gcflags=$(if $(VNET_DEBUG_yes),-N -l)
go-wip:
	$(gobuild) -o $@ $(reltopdir)wip/y
