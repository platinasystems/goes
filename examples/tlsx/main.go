// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This command provides a TLS network in which an exchange bridges connected
// hosts. e.g.
//
//	$ tlsx -state ~/.local/state/tlsx/x create-cert
//	$ tlsx -state ~/.local/state/tlsx/h1 create-cert
//	$ tlsx -state ~/.local/state/tlsx/h2 create-cert
//	$ sudo ip netns exec x tlsx -state ~/.local/state/tlsx/x \
//		start exchange bridge leasing 10.200.1.0/30 10.200.1.2 &
//	$ sudo ip netns exec h1 tlsx -state ~/.local/state/tlsx/h1 \
//		start tap -prefix 10.200.1.1/30 x &
//	$ sudo ip netns exec h2 tlsx -state ~/.local/state/tlsx/h2 \
//		start tap x &
//	$ sudo ip netns exec h2 ping -c 1 10.200.1.1
//	...
//	$ sudo ip netns exec h1 ping -c 1 10.200.1.2
//	...
//
// In the above, the exchange, `x` provides dynamic addresses w/in the prefix
// 10.200.1.0/30 starting at 10.200.1.2;  host `h1` joins the exchange with
// static address 10.200.1.1 and host `h2` joins requesting a dynamic address.
package main

import (
	"github.com/platinasystems/goes/v2/pkg/coreutils"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/service"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/net/udpecho"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/termination"
)

func main() {
	tlsx.Daemons["udp"] = map[string]any{
		"echo": udpecho.Reply,
	}
	tlsx.Root["daemon"] = tlsx.Daemons
	tlsx.Show["build"] = program.Build
	tlsx.Show["main"] = program.Main
	tlsx.Root["show"] = tlsx.Show
	tlsx.Root["udp"] = map[string]any{
		"ping": udpecho.Ping,
	}
	tlsx.Root["command"] = coreutils.Command
	tlsx.Root["standby"] = termination.Standby
	tlsx.Root["start"] = goes.Start
	service.Selection["cat"] = coreutils.Cat
	service.Selection["command"] = coreutils.Command
	service.Selection["echo"] = coreutils.Echo
	goes.Root = tlsx.Root
	goes.Reload = certs.Subscribers.Invalidate
	goes.Main()
}
