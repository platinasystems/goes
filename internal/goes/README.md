Package `goes`, combined with a compatibly configured Linux kernel, provides a
monolithic embedded system.

*How is this pronounced?*

We prefer "**go e-s**", aka. "**go e**[_mbedded_]**-s**[_ystem_]".

*How is it monlithic?*

The [Live CD] of most Linux distributions is an assembly of binary packages and
base configuration. It may also include a guided self installer that copies its
contents to the target along with other network acquired packagesfrom. These
binary packages are built from interdependent source packages by a maintainer.

With `goes`, the package assembly, base configuration, and interdependency
is handled by `go build`.  The result is a single, thus monolithic, program
including all of the commands that the maintainer intends to support on the
target machine.  This may include a guided self installer that, generally,
doesn't require network install of anything else.

Alternatively, `goes` may run as a self-spawning daemons and interactive sub-
commands within a minimal Linux distribution.

*What are machines?*

Machines are main packages that provide a `goes` command manifest,
configuration, and customization before calling `Goes.Main()`.
See [main] for examples.

To build the example machine.

```console
$ make goes-example
go generate ./copyright
go generate ./version
go build -o goes-example ./main/goes-example
```

To build all machines,

```console
$ make
go generate ./copyright
go generate ./version
...
```

To install,

```console
$ sudo ./goes-example install
$ goes show-commands
...
```

To enable BASH completion after install,

```console
. /usr/share/bash-completion/completions/goes
```

To run commands without install,

```console
$ ./goes-example show-commands
...
```

To debug,

```console
$ gdb ./goesd-example
```

Each [goes/cmd] provides _apropos_, _completion_, _man_, and _usage_.
The command may also provide context sensitive _help_, _README_, and _godoc_.

- `goes apropos` _COMMAND_
- `goes complete` _COMMAND_ [_ARGS_]...
- `goes man` _COMMAND_
- `goes usage` _COMMAND_
- `goes help` _COMMAND_ [_ARGS_]...
- https://github.com/platinasystems/go/tree/master/internal/goes/cmd/COMMAND/README.md
- https://godoc.org/github.com/platinasystems/go/internal/goes/cmd/COMMAND

---

*&copy; 2015-2017 Platina Systems, Inc. All rights reserved.
Use of this source code is governed by this BSD-style [LICENSE].*

[main]: ../../main
[goes/cmd]: ./cmd
[LICENSE]: ../../LICENSE
[Live CD]: https://en.wikipedia.org/wiki/Live_CD
