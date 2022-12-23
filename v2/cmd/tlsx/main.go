// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This command provides a TLS network in which an exchange bridges connecting
// hosts.
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
	daemon.Exchange.Selection = map[string]any{
		"approve":     admin.Func,
		"build":       program.Build,
		"cat":         cat.Func,
		"command":     command.Func,
		"deny":        admin.Func,
		"echo":        echo.Func,
		"main":        program.Main,
		"subscribers": certs.Subscribers,
		"tenants":     &daemon.Exchange.Bridge.Leasing,
	}
	goes.Root = map[string]any{
		"build":         program.Build,
		"cert":          certs.Self,
		"create-cert":   create_cert.Func,
		"daemon":        daemon.Func,
		"exec":          exec.Func,
		"main":          program.Main,
		"port":          port.Map,
		"rpc":           service.Request{tlsx.RPC}.Func,
		"runtime-dir":   xdg.RunTimeDir,
		"state-dir":     dir.Name,
		"start":         start.Func,
		"subscribe":     subscribe.Func,
		"subscribers":   certs.Subscribers,
		"subscriptions": certs.Subscriptions,
	}
	goes.Main()
}
