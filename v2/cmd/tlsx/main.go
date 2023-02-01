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
	"github.com/platinasystems/goes/v2/pkg/goes/service"
	"github.com/platinasystems/goes/v2/pkg/goes/start"
	"github.com/platinasystems/goes/v2/pkg/goes/tlsx/admin"
	"github.com/platinasystems/goes/v2/pkg/goes/tlsx/create_cert"
	"github.com/platinasystems/goes/v2/pkg/goes/tlsx/daemon"
	"github.com/platinasystems/goes/v2/pkg/goes/tlsx/exec"
	"github.com/platinasystems/goes/v2/pkg/goes/tlsx/subscribe"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/port"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/dir"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/xdg"
)

func main() {
	var starter any = start.Func
	if program.IsKoApp() {
		starter = daemon.Func
	}
	daemon.Exchange.Selection = map[string]any{
		"approve": admin.Func,
		"cat":     cat.Func,
		"command": command.Func,
		"deny":    admin.Func,
		"echo":    echo.Func,
		"show": map[string]any{
			"build":       program.Build,
			"main":        program.Main,
			"subscribers": certs.Subscribers,
			"tenants":     &daemon.Exchange.Bridge.Leasing,
		},
	}
	goes.Root = map[string]any{
		"approve":     exec.IPC,
		"command":     command.Func,
		"create-cert": create_cert.Func,
		"daemon":      daemon.Func,
		"deny":        exec.IPC,
		"exec":        exec.Func,
		"port":        port.Map,
		"rpc":         service.Request{tlsx.RPC}.Func,
		"runtime-dir": xdg.RunTimeDir,
		"show": map[string]any{
			"build": program.Build,
			"cert": map[string]any{
				"pem":  certs.Self.MarshalPEM,
				"text": certs.Self.MarshalText,
			},
			"main":          program.Main,
			"state-dir":     dir.Name,
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
