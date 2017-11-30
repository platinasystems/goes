#!/usr/bin/make

ALL = ip ip.test

include $(shell git rev-parse --show-cdup)go.mk

ip:
	$(gobuild)

ip.test:
	$(gotest) -c
