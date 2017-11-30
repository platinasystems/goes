#!/usr/bin/make

ALL = goes-platina-mk2-lc1-bmc

include $(shell git rev-parse --show-cdup)go.mk

goes-platina-mk2-lc1-bmc: arch=arm
goes-platina-mk2-lc1-bmc: tags=netgo$(diag_tag)
goes-platina-mk2-lc1-bmc: ldflags=-d
goes-platina-mk2-lc1-bmc:
	$(gobuild)
