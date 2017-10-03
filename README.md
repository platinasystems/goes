This repos contains Platina System's open source GO projects.

You'll need `platinasystems/fe1` and `platinasystems/firmware-fe1a` to compile
`goes-platina-mk1` or `go-wip`.

```console
$ git clone git@github.com:platinasystems/fe1 ../fe1
$ git clone git@github.com:platinasystems/firmware-fe1a ../firmware-fe1a
```

Alternatively, you may build `goes-platina-mk1` to plugin an existing
`/usr/lib/goes/fe1.so`.

To install a select MACHINE,

```console
$ make -B goes-MACHINE
$ sudo ./goes-MACHINE install
```

Some machines also have a self extracting, compressed archive installer.

```console
$ make -B goes-MACHINE-installer
$ sudo ./goes-MACHINE-installer
```

These are the available MACHINEs,

- [example] \(GOARCH: amd64 or armhf)
- [platina-mk1] \(GOARCH: amd64)
- [platina-mk1-bmc] \(GOARCH: armhf)

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
[example]: goes/goes-example/README.md
[platina-mk1]: goes/goes-platina-mk1/README.md
[errata]: docs/Errata.md
