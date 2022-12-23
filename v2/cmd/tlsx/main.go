// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This command provides a TLS network in which an exchange bridges connecting
// hosts.
package main

import (
	"github.com/platinasystems/goes/v2/pkg/goes/cat"
	"github.com/platinasystems/goes/v2/pkg/goes/command"
	"github.com/platinasystems/goes/v2/pkg/goes/echo"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/goes/service"
	"github.com/platinasystems/goes/v2/pkg/goes/show"
	"github.com/platinasystems/goes/v2/pkg/goes/start"
	"github.com/platinasystems/goes/v2/pkg/goes/tlsx/admin"
	"github.com/platinasystems/goes/v2/pkg/goes/tlsx/create_cert"
	"github.com/platinasystems/goes/v2/pkg/goes/tlsx/daemon"
	"github.com/platinasystems/goes/v2/pkg/goes/tlsx/exec"
	"github.com/platinasystems/goes/v2/pkg/goes/tlsx/subscribe"
	"github.com/platinasystems/goes/v2/pkg/goes/xdg"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/port"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/dir"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

const Usage = `
usage:	{{.Prog}} [<options>] address [<exchange> [<exchange-address>]]
		Assign or print exchange address.
	{{.Prog}} [<options>] create-cert
		Create key and certifcate for host, consumer, or exchange.
	{{.Prog}} [<options>] exec <exchange> <command> [<args>]
		Run command on exchange.
	{{.Prog}} json ...
		JSON format object.
	{{.Prog}} show ...
		Text format object.
	{{.Prog}} [<options>] rpc <host> <func> [<args>]
		Blocking RPC to <host>.
	{{.Prog}} [<options>] start <routines>
		Ordered start of one or more daemon go-routines.
	{{.Prog}} [<options>] subscribe <registry-address>
		Request exchange service.
{{print .Flags}}
  <address>	A network address and port, e.g.
	- :8003
	- 127.0.0.1:8003
	- [::1]:8003
	- unix://PATH
	- unix://@NAME

  <exchange>, <guest>
	The primary DNS name or subject-key-id of a certificate.

  <routine>	One or more, ordered daemon go-routines, e.g.
	- exchange [bridge [leasing <network> <base>]]
	- registry
	- rpc
	- tap [-unix <#>] [-prefix <prefix>] <exchange>
`

func main() {
	selection.Usage = Usage
	tenants := &daemon.Exchange.Bridge.Leasing
	showprog := selection.Map{
		"build": selection.Map{
			"id":   show.Text{program.Build.Id}.Func,
			"info": show.Text{program.Build.Info}.Func,
		}.Select,
		"main": selection.Map{
			"reference": show.Text{program.Main.Reference}.Func,
			"version":   show.Text{program.Main.Version}.Func,
		}.Select,
	}.Select
	daemon.Exchange.Selector = selection.Map{
		"approve": admin.Func,
		"cat":     cat.Func,
		"command": command.Func,
		"deny":    admin.Func,
		"echo":    echo.Func,
		"show": selection.Map{
			"json": selection.Map{
				"tenants": show.JSON{tenants}.Func,
			}.Select,
			"program":     showprog,
			"subscribers": show.Text{certs.Subscribers}.Func,
			"tenant":      show.KeyText{tenants}.Func,
			"tenants":     show.Text{tenants}.Func,
		}.Select,
	}
	selection.Map{
		"create-cert": create_cert.Func,
		"daemon":      daemon.Func,
		"exec":        exec.Func,
		"rpc":         service.Request{tlsx.RPC}.Func,
		"show": selection.Map{
			"program":       showprog,
			"cert":          show.Text{certs.Self}.Func,
			"exchange-port": show.TextContext{port.Exchange}.Func,
			"registry-port": show.TextContext{port.Registry}.Func,
			"rpc-port":      show.TextContext{port.RPC}.Func,
			"subscribers":   show.Text{certs.Subscribers}.Func,
			"subscriptions": show.Text{certs.Subscriptions}.Func,
			"state":         show.Text{dir.Cached.Name}.Func,
			"xdg":           xdg.Show,
		}.Select,
		"start":     start.Func,
		"subscribe": subscribe.Func,
	}.Main()
}
