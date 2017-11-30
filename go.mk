#!/usr/bin/make
# make V=1 for verbose go builds
# make diag=yes for enhanced diagnostics

# refresh package.go at every make
_gen := $(shell cd $(shell git rev-parse --show-cdup) && go generate)

diag_yes:=$(filter yes,$(diag))
diag_tag:=$(if $(diag_yes), diag)

gobuild = $(if $(arch),env GOARCH=$(arch) )go build$(if $(tags),\
-tags "$(tags)")$(if $(gcflags),\
-gcflags "$(gcflags)")$(if $(ldflags),\
-ldflags "$(ldflags)")$(if $(V),\
-v)

gotest = $(if $(arch),env GOARCH=$(arch) )go test$(if $(tags),\
-tags "$(tags)")$(if $(gcflags),\
-gcflags "$(gcflags)")$(if $(ldflags),\
-ldflags "$(ldflags)")$(if $(V),\
-x)

.PHONY: all
all: $(ALL)

.PHONY: clean
clean:
	@rm -f $(ALL) *.so *.zip
	@git clean -d -f
