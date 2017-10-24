// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import "github.com/platinasystems/go/goes/cmd/ip/internal/options"

func New(args []string) (*options.Options, []string) {
	opt, args := options.New(args)
	args = opt.Parms.More(args,
		"address",
		"broadcast",
		"index",
		"link",
		"mtu",
		"name",
		"numrxqueues",
		"numtxqueues",
		[]string{"txqueuelen", "txqlen"},
	)
	return opt, args
}
