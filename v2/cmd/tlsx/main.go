// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This command provides a TLS network in which an exchange bridges connecting
// hosts. e.g.
//
//	$ sudo ip netns exec x ./tlsx -state ~/t/x -verbose start \
//		exchange bridge leasing 10.200.1.0/30 10.200.1.2 &
//	$ sudo ip netns exec h1 ./tlsx -state ~/t/h1 -verbose start \
//		tap -prefix 10.200.1.1/30 x &
//	$ sudo ip netns exec h2 ./tlsx -state ~/t/h2 -verbose start tap x &
//	$ sudo ip netns exec h2 ping -c 1 10.200.1.1
//	...
//	$ sudo ip netns exec h1 ping -c 1 10.200.1.2
//	...
package main

import (
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/goes/cat"
	"github.com/platinasystems/goes/v2/pkg/goes/command"
	"github.com/platinasystems/goes/v2/pkg/goes/echo"
	"github.com/platinasystems/goes/v2/pkg/goes/start"
	"github.com/platinasystems/goes/v2/pkg/goes/tlsx/create_cert"
	"github.com/platinasystems/goes/v2/pkg/goes/tlsx/exec"
	"github.com/platinasystems/goes/v2/pkg/goes/tlsx/subscribe"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/exchange"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/service"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/dir"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/tap"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/xdg"
)

func main() {
	var starter any = start.Func
	daemons := map[string]any{
		"exchange": exchange.Daemon,
		"tap":      tap.Daemon,
	}
	if program.IsKoApp() {
		// start daemons directly instead of through detached child
		starter = daemons
	}
	service.Selection["cat"] = cat.Func
	service.Selection["command"] = command.Func
	service.Selection["echo"] = echo.Func
	goes.Root = map[string]any{
		"approve":     exec.IPC,
		"command":     command.Func,
		"create-cert": create_cert.Func,
		"daemon":      daemons,
		"deny":        exec.IPC,
		"exec":        exec.Func,
		"show": map[string]any{
			"build": program.Build,
			"cert": map[string]any{
				"pem":  certs.Self.MarshalPEM,
				"text": certs.Self.MarshalText,
			},
			"dir": map[string]any{
				"runtime": xdg.RunTimeDir,
				"state":   dir.Name,
			},
			"main":          program.Main,
			"registry":      exec.IPC,
			"subscribers":   certs.Subscribers,
			"subscriptions": certs.Subscriptions,
			"tenants":       exec.IPC,
		},
		"start":     starter,
		"subscribe": subscribe.Func,
	}
	goes.Reload = certs.Subscribers.Invalidate
	goes.Main()
}
