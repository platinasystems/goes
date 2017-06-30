// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import (
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
)

var (
	Family = []string{"-4", "-6", "-B", "-M", "-0"}
	Flags  = []interface{}{
		[]string{"-human", "-human-readable"},
		[]string{"-s", "-stats", "-statistics"},
		[]string{"-d", "-details"},
		[]string{"-o", "-oneline"},
		[]string{"-r", "-resolve"},
		[]string{"-a", "-all"},
		[]string{"-c", "-color"},
		[]string{"-t", "-timestamp"},
		[]string{"-ts", "-tshort"},
		"-iec",
	}
	Parms = []interface{}{
		[]string{"-l", "-loops"},
		[]string{"-f", "-family"},
		[]string{"-rc", "-rcvbuf"},
	}
)

type Options struct {
	Flags *flags.Flags
	Parms *parms.Parms
}

// Parse common IP options from command arguments.
func New(args []string) (*Options, []string) {
	opt := new(Options)
	opt.Flags, args = flags.New(args, Flags...)
	opt.Parms, args = parms.New(args, Parms...)
	family, args := flags.New(args, Family)
	switch {
	case family.ByName["-4"]:
		opt.Parms.ByName["-f"] = "inet"
	case family.ByName["-6"]:
		opt.Parms.ByName["-f"] = "inet6"
	case family.ByName["-B"]:
		opt.Parms.ByName["-f"] = "bridge"
	case family.ByName["-M"]:
		opt.Parms.ByName["-f"] = "mpls"
	case family.ByName["-0"]:
		opt.Parms.ByName["-f"] = "link"
	}
	return opt, args
}
