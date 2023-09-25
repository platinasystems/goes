Package `goes` provides a monolithic program and configuration for embedded
for Unix systems.

*How is this pronounced?*

We prefer "**go e-s**", aka. "**go e**[_mbedded_]**-s**[_ystem_]".

*How is it monlithic?*

The [Live CD image] of Unix distributions are generally an assembly of binary
packages and base configuration that may also include a guided self installer
that copies its contents to the target along with other network acquired
packages. The binary packages are built from interdependent source packages by
maintainers.

With `goes`, the package assembly, base configuration, and interdependency
is handled by `go build`.  The result is a single, thus monolithic, program
including all of the commands that the maintainer intends to support on the
target machine.  This may include a guided self installer that, generally,
doesn't require network install of anything else.

Alternatively, `goes` may run as a self-spawning daemons and interactive sub-
commands within a minimal base Unix image.

*What are `goes` machines?*

Machines are main packages that assemble a root function map before calling
`pkg/goes.Main()` that selects the mapped function with the program's
arguments. The top level [main] is the primary example machine with others
provided in [examples]. Alternatively, the main of an external machine package
may import the GPL-v2 [LICENSE] packages in [pkg] beginning with [pkg/goes].

The top-level, primary example program is built with,

```console
$ [GOOS=<goos>] [GOARCH=<goarch>] go build
```

Or build and assemble a minimal, multi-platform container image with,

```console
$ ko build [-B] [-L] --platform=all
```

To run a specific platform,

```console
$ export DOCKER_DEFAULT_PLATFORM=linux/<goarch>
$ docker run --rm ko.local/goes ...
```


---

*&copy; 2015-2023 Platina Systems, Inc. All rights reserved.
Use of this source code is governed by this BSD-style [LICENSE].*

[LICENSE]: ./LICENSE
[examples]: ./examples
[main]: ./main.go
[pkg]: ./pkg
[pkg/goes]: ./pkg/goes
[Live CD]: https://en.wikipedia.org/wiki/Live_CD
