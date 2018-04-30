#!/usr/bin/make
# make plugin=yes|no to force plugin enable/disable
# make VNET_DEBUG=yes to enable vnet debugging checks and flags for gdb.

fe1_pkg  := github.com/platinasystems/fe1
fe1_dir  := $(shell go list -e -f {{.Dir}} $(fe1_pkg))
_fe1_gen := $(if $(fe1_dir),$(shell cd $(fe1_dir) && go generate))
libfe1so := $(wildcard /usr/lib/goes/fe1.so)

fe1a_pkg := github.com/platinasystems/firmware-fe1a
fe1a_dir := $(shell go list -e -f {{.Dir}} $(fe1a_pkg))
_fe1a_gen := $(if $(fe1a_dir),$(shell cd $(fe1a_dir) && go generate))

plugin := $(if $(fe1_dir),no,yes)
plugin_yes:=$(filter yes,$(plugin))
plugin_tag:=$(if $(plugin_yes), plugin)
plugin_fe1:=$(if $(plugin_yes),fe1.so)

VNET_DEBUG_yes:=$(filter yes,$(VNET_DEBUG))
VNET_DEBUG_tag:=$(if $(VNET_DEBUG_yes), debug)

fe1.so: tags=vfio$(VNET_DEBUG_tag)$(diag_tag)
fe1.so: $(if $(fe1_dir),| $(fe1_gen) $(fe1a_gen),$(libfe1so))
	$(if $(fe1_dir),$(gobuild) -buildmode=plugin ./main/fe1,\
		$(if $(libfe1so),cp $(libfe1so),touch) $@)

$(fe1_gen): $(addprefix $(fe1_dir)/,LICENSE PATENTS)
	go generate $(fe1_pkg)

$(fe1a_gen): $(fe1a_dir)/LICENSE
	go generate $(fe1a_pkg)
