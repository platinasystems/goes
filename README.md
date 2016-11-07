This repos contains Platina System's open source GO projects.

To install a select MACHINE,

```console
$ make
$ sudo ./goes-MACHINE install
```

These are the available MACHINEs,

- [example] \(GOARCH: amd64 or arm)
- [platina-mk1] \(GOARCH: amd64)

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

---

*&copy; 2015-2016 Platina Systems, Inc. All rights reserved.
Use of this source code is governed by this BSD-style [LICENSE].*

[LICENSE]: LICENSE
[example]: goes/goes-example/README.md
[platina-mk1]: goes/goes-platina-mk1/README.md
