#!/usr/bin/make

ALL = goes-platina-mk1-bmc

include $(shell git rev-parse --show-cdup)go.mk

goes-platina-mk1-bmc: arch=arm
goes-platina-mk1-bmc: tags=netgo$(diag_tag)
goes-platina-mk1-bmc: ldflags=-d
goes-platina-mk1-bmc:
	$(gobuild)
