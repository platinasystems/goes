This repos contains Platina System's open source GO projects.

First fetch and install the goes builder,

```console
$ go get github.com/platinasystems/go/main/goes-build
```

You'll need `platinasystems/fe1` and `platinasystems/firmware-fe1a` to build
`goes-platina-mk1`.

```console
$ git clone git@github.com:platinasystems/fe1 \
	GOPATH/src/github.com/platinasystems/fe1
$ git clone git@github.com:platinasystems/firmware-fe1a 
	GOPATH/src/github.com/platinasystems/firmware-fe1a
```

Alternatively, you may build `goes-platina-mk1` to plugin an existing
`/usr/lib/goes/fe1.so`.

To install a select MACHINE,

```console
$ goes-build goes-MACHINE
$ sudo ./goes-MACHINE install
```

Some machines also have a self extracting, compressed archive installer.

```console
$ goes-build goes-MACHINE-installer
$ sudo ./goes-MACHINE-installer
```

These are the available machines,

- [example] \(GOARCH: amd64 or armhf)
- [boot] \(GOARCH: amd64)
- [platina-mk1] \(GOARCH: amd64)
- [platina-mk1-bmc] \(GOARCH: armhf)
- [platina-mk2-lc1-bmc] \(GOARCH: armhf)
- [platina-mk2-mc1-bmc] \(GOARCH: armhf)

To stop and remove,

```console
$ sudo goes uninstall
```

To enable BASH completion after install,

```console
. /usr/share/bash-completion/completions/goes
```

To see the commands available on the installed MACHINE,

```console
$ goes help
```

Or,

```console
$ goes
goes> help
```

Then `man` any of the listed commands or `man cli` to see how to use the
command line interface.

See also [errata].

---

*&copy; 2015-2016 Platina Systems, Inc. All rights reserved.
Use of this source code is governed by this BSD-style [LICENSE].*

[LICENSE]: LICENSE
[errata]: docs/Errata.md
[example]: main/goes-example/README.md
[boot]: main/goes-boot/README.md
[platina-mk1]: main/goes-platina-mk1/README.md
[platina-mk1-bmc]: main/goes-platina-mk1-bmc/README.md
[platina-mk2-lc1-bmc]: main/goes-platina-mk2-lc1-bmc/README.md
[platina-mk2-mc1-bmc]: main/goes-platina-mk2-mc1-bmc/README.md
