#!/usr/bin/make

ALL = goes-example goes-example-arm goes-example.test

include $(shell git rev-parse --show-cdup)go.mk

goes-example:
	$(gobuild)

goes-example.test:
	$(gotest) -c

goes-example-arm: arch=arm
goes-example-arm: tags=netgo
goes-example-arm: ldflags=-d
goes-example-arm:
	$(gobuild) -o $@
