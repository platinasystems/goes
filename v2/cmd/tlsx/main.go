// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This command provides a TLS network in which an exchange bridges connecting
// hosts.
package main

import (
	"github.com/platinasystems/goes/v2/pkg/goes/daemon"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/goes/show"
	"github.com/platinasystems/goes/v2/pkg/goes/tlsx"
	"github.com/platinasystems/goes/v2/pkg/goes/tlsx/tap"
	"github.com/platinasystems/goes/v2/pkg/goes/xdg"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

const Usage = `
usage:	{{.Prog}} [<options>] address [<exchange> [<exchange-address>]]
		Assign or print exchange address.
	{{.Prog}} [<options>] create-cert
		Create key and certifcate for host, consumer, or exchange.
	{{.Prog}} [<options>] [exec] <exchange> approve [<guest(s)>]
		Approve subscriptions.
	{{.Prog}} [<options>] [exec] <exchange> subscribers
		List certificates of exchange subscribers.
	{{.Prog}} [<options>] [exec] <exchange> deny [<guest(s)>]
		Deny client(s) subscription.
	{{.Prog}} [<options>] [exec] <exchange> [<command> [<args>]]
		Run request on host connected through exchange.
	{{.Prog}} show build-id
	{{.Prog}} show build-info
	{{.Prog}} [<options>] show cert
		Print local certificate.
	{{.Prog}} [<options>] show subscribers
		List certificates of exchanges.
	{{.Prog}} [<options>] show subscriptions
		List exchange certificates.
	{{.Prog}} show main-reference
		PATH@SYMVER
	{{.Prog}} show version
		SYMVER
	{{.Prog}} [<options>] start exchange [-r <address>] [-x <address>]
		Start TLS exchange service.
	{{.Prog}} [<options>] start tap [-u <unit>] [<exchange>]
		Start link tunnel.
	{{.Prog}} [<options>] subscribe <registry-address>
		Request exchange service.
{{print .Flags}}
  <address>
	A network address and port, e.g.
		:8003
		[::1]:8003
		unix://PATH
		unix://@NAME
  <exchange>, <guest>
	The primary DNS name or subject-key-id of a certificate.
`

func main() {
	selection.Usage = Usage
	selection.Map{
		"address":     tlsx.Address,
		"create-cert": tlsx.CreateCert,
		"daemon": selection.Map{
			"exchange": tlsx.Exchange,
			tap.Key:    tap.Daemon,
		}.Select,
		"exec": tlsx.Exec,
		"show": selection.Map{
			"build-id":       show.New(program.BuildId),
			"build-info":     show.New(program.BuildInfo),
			"cert":           tlsx.ShowCert,
			"main-reference": show.New(program.MainReference),
			"subscribers":    tlsx.ShowSubs,
			"subscriptions":  tlsx.ShowSubs,
			"state":          show.New(state.Cache.Dir),
			"xdg":            xdg.Show,
			"version":        show.New(program.MainVersion),
		}.Select,
		"start": selection.Map{
			"exchange": daemon.Start,
			tap.Key:    daemon.Start,
		}.Select,
		"subscribe": tlsx.Subscribe,
	}.Main(func(m selection.Map) error {
		for _, x := range tlsx.Exchanges() {
			m[x] = tlsx.Exec
		}
		return nil
	})
}
