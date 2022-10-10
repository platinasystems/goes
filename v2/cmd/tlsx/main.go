// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

/*
The tlsx provides these TLS network roles: host, consumer, and exchange.

Usage:
	tlsx [-d DIR] alias [<name> [<subject-key-id>]]
		Add or print subject-key-id association.
	tlsx [-d DIR] authorize [<subject-key-id(s)>]
		Authorize host consumer.
	tlsx [-d DIR] create-cert
		Create key and certifcate for host, consumer, or exchange.
	tlsx [-d DIR] exchange approve [<subject-key-id(s)>]
		Approve client(s) subscription.
	tlsx [-d DIR] exchange clients
		List certificates of exchange clients.
	tlsx [-d DIR] exchange deny [<subject-key-id(s)>]
		Deny client(s) subscription.
	tlsx [-d DIR] exchange start [-r PORT] [-x PORT]
		Start TLS exchange service.
	tlsx [-d DIR] host
		Service consumer requests through subscribed exchange(s).
	tlsx [-d DIR] jump <host> [<request> [<args>]]
		Run request on host connected through exchange.
	tlsx [-d DIR] revoke [<subject-key-id(s)>]
		Revoke consumer authorization.
	tlsx show build-id
	tlsx show build-info
	tlsx [-d DIR] show cert
		Print local certificate.
	tlsx [-d DIR] show clients
		List client certificates of exchange.
	tlsx [-d DIR] show exchanges
		List exchange certificates of host or consumer.
	tlsx show main-reference
		PATH@SYMVER
	tlsx show version
		SYMVER
	tlsx [-d DIR] subscribe
		Register host or consumer with exchange.
*/
package main

import (
	"flag"

	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/goes/tlsx"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state"
)

func main() {
	flag.String("state", state.DirDefault.String(), "")
	flag.Bool("verbose", false, "")
	selection.Map{
		"alias":       tlsx.Alias,
		"authorize":   tlsx.AuthorizeOrRevoke,
		"create-cert": tlsx.CreateCert,
		"exchange":    tlsx.Exchange.Select,
		"host":        tlsx.Host,
		"jump":        tlsx.Jump,
		"revoke":      tlsx.AuthorizeOrRevoke,
		"show":        tlsx.Show.Select,
		"subscribe":   tlsx.Subscribe,
	}.Main()
}
