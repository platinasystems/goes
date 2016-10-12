Package `goes`, combined with a compatibly configured Linux kernel, provides a
monolithic embedded system.

*How is this pronounced?*

We prefer "*go e-s*".

*How is it monlithic?*

The [Live CD] of most Linux distributions is an assembly of binary packages and
base configuration. The Live CD may also include a guided self installer that
copies the contained binary packages to the target along with other packages
acquired from network repositories. These binary packages are built from
interdependent source packages by the distro maintainer.

With `goes`, the package assembly, base configuration, and interdependency
is handled by `go build`.  The result is a single, thus monolithic, program
including all of the commands that the maintainer intends to support on the
target machine.  This may include a guided self installer but doesn't generally
require net install of anything else.

Alternatively, `goes` may run as a self-spawning daemons within a debian based
system. In this mode the same daemons built within the `goes` monolith are run
from the internal master control daemon installed as `/usr/sbin/goesd`.

*What are machines?*

Machines are main packages that contain register a command manifest,
configuration, and customization before calling `goes.Goes`. Examples may be
found in the `example` and `platina` directories.

Run this to build the example debian daemon.

```console
$ go build -o goesd-example ./example
```

You may also run on the  build host with,

```console
$ ./goesd-example
```

And of course,

```console
$ gdb ./goesd-example
```

Run this to rebuild the example as a static embedded machine.

```console
$ go build -o goes-example -tags netgo -a -ldflags -d ./example

```

Then run this to make a linux initrd.

```console
tmp=tmp$$
install -s -D goes-example ${tmp}/init
lnstall -d ${tmp}/bin
ln -sf ../init ${tmp}/bin/goes
(cd ${tmp} && find . | cpio --quiet -H newc -o --owner 0:0) |
	xz -z --check=crc32 -9 -c > goes-example.cpio.xz
```

*What is a goes base /init?*

See `coreutils/slashinit`.
This mounts then pivots to the target root and runs it's `/sbin/init`.

`goes` includes this `example/init`.

*What about /sbin/init and /usr/sbin/goesd?*

See `coreutils/sbininit`.
Both of these start a redis server then runs all embedded daemons.
`/sbin/init` returns to `goes.Goes` to start a console shell whereas
`/usr/sbin/goesd` just waits for a signal to kill or restart the daemons.

---

*&copy; 2015-2016 Platina Systems, Inc. All rights reserved.
Use of this source code is governed by this BSD-style [LICENSE].*

[github.com/platinasystems/goesdeb]: https://github.com/platinasystems/goesdeb
[LICENSE]: ../LICENSE
[Live CD]: https://en.wikipedia.org/wiki/Live_CD
[import]: https://golang.org/ref/spec#Import_declarations
[blank]: https://golang.org/ref/spec#Blank_identifier
[baseboard management controller]: https://en.wikipedia.org/wiki/Intelligent_Platform_Management_Interface#Baseboard_management_controller
[naked switch data plane]: https://github.com/platinasystems/rfc/blob/master/nsdp.md
