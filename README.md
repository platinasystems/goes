When imported by a machine main, package `goes` provides a monolithic embedded
distro for Linux systems,

*How is this pronounced?*

We prefer "**go e-s**", aka. "**go e**[_mbedded_]**-s**[_ystem_]".

*How is it monlithic?*

The [Live CD] of most Linux distributions is an assembly of binary packages and
base configuration. It may also include a guided self installer that copies its
contents to the target along with other network acquired packages. These binary
packages are built from interdependent source packages by a maintainer.

With `goes`, the package assembly, base configuration, and interdependency
is handled by `go build`.  The result is a single, thus monolithic, program
including all of the commands that the maintainer intends to support on the
target machine.  This may include a guided self installer that, generally,
doesn't require network install of anything else.

Alternatively, `goes` may run as a self-spawning daemons and interactive sub-
commands within a minimal Linux distribution.

*What are machines?*

Machines are main packages that provide a `goes` command manifest,
configuration, and customization before calling `Goes.Main()`.  See
`https://github.com/platinasystems/goes-MACHINE` for examples.

These machines are available at https://github.com/platinasystems/,

- [goes-example] \(GOARCH: amd64 or armhf)
- [goes-boot] \(GOARCH: amd64)
- [goes-platina-mk1] \(GOARCH: amd64)
- [goes-bmc] \(GOARCH: armhf)

Most machines are built with just `go build` but some, like the _bmc_, may also
use [goes-build] to build a properly configured kernel and embed itself as an
_initrd_.

To install,

```console
$ sudo ./goes-MACHINE install
...
```

To stop and remove,

```console
$ sudo goes uninstall
```

To enable BASH completion after install,

```console
. /usr/share/bash-completion/completions/goes
```

To run commands without install,

```console
$ ./goes-MACHINE COMMAND [ARGS]...
...
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

To debug,

```console
$ gdb ./goes-MACHINE
```

Each [goes/cmd] provides _apropos_, _completion_, _man_, and _usage_.
The command may also provide context sensitive _help_, _README_, and _godoc_.

- `goes apropos` _COMMAND_
- `goes complete` _COMMAND_ [_ARGS_]...
- `goes man` _COMMAND_
- `goes usage` _COMMAND_
- `goes help` _COMMAND_ [_ARGS_]...
- https://github.com/platinasystems/goes/blob/master/cmd/COMMAND/README.md
- https://godoc.org/github.com/platinasystems/goes/cmd/COMMAND

See also [errata].

---

*&copy; 2015-2019 Platina Systems, Inc. All rights reserved.
Use of this source code is governed by this BSD-style [LICENSE].*

[LICENSE]: ../LICENSE
[errata]: docs/Errata.md
[goes/cmd]: ./cmd
[goes-example]: https://github.com/platinasystems/goes-example.git
[goes-boot]: https://github.com/platinasystems/goes-boot.git
[goes-platina-mk1]: https://github.com/platinasystems/goes-platina-mk1.git
[goes-bmc]: https://github.com/platinasystems/goes-bmc.git
[goes-build]: https://github.com/platinasystems/goes-build.git
[Live CD]: https://en.wikipedia.org/wiki/Live_CD
