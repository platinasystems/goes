#!/usr/bin/make

subdirs := main/ip
subdirs += main/goes-example
subdirs += main/goes-coreboot
subdirs += main/goes-platina-mk1
subdirs += main/goes-platina-mk1-bmc
subdirs += main/goes-platina-mk2-lc1-bmc
subdirs += main/goes-platina-mk2-mc1-bmc

.PHONY: all
all:
	@$(foreach subdir,$(subdirs),$(MAKE) -C $(subdir) && ):

.PHONY: clean
clean:
	@$(foreach subdir,$(subdirs),$(MAKE) -C $(subdir) clean && ):
