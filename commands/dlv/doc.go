// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

/*
Package goes/dlv provides command wrappers to Derek Parker's delve
debugger.  Use `dlv` with a locally available source tree; otherwise,
use `dlvd` to start a headless server an connect to it from the source
machine.

	goes> dlv COMMAND [ARGS]...
	goes> dlvd [-l ADDR:PORT] COMMAND [ARGS]...

If COMMAND is a daemon, these will `dlv attach PID` with the PID recorded
in `/run/goes/pids/DAEMON`; otherwise, these
`dlv exec /usr/bin/goes COMMAND [ARGS]...` and break at COMMAND's Main.

Daemons may be configured to standby at start through this redis config in
`/etc/goes/goesd`.

	hset -q standby DAEMON true

When running `dlv DAEMON`, first set a breakpoint then signal the daemon to
clear standby with this.

	$ sudo kill -CONT $(cat /run/goes/pids/DAEMON)

Please consider building with -gcflags="-N -l" to disable optimization
and inlines.
*/
package dlv
