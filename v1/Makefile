#!/usr/bin/make -f

maintainer := $(shell git config --get user.name)
maintainer += <$(shell git config --get user.email)>
signingkey := $(shell git config --get user.signingkey)

all: ; @:

install:
	@:

clean:
	@:

distclean:
	rm -f debian/debhelper-build-stamp debian/files debian/*.substvars
	rm -rf debian/.debhelper debian/goes-completion

bindeb-pkg:
	debuild -b\
		-k"$(signingkey)"\
		-m"$(maintainer)"\
		--lintian-opts --profile debian

.PHONY: all install bindeb-pkg clean distclean
