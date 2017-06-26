// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import (
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
)

// Parse IP options
func New(args []string) (flags.Flag, parms.Parm, []string) {
	flag, args := flags.New(args,
		"-human", "-human-readable",
		"-s", "-stats", "-statistics",
		"-d", "-details",
		"-o", "-oneline",
		"-r", "-resolve",
		"-a", "-all",
		"-c", "-color",
		"-t", "-timestamp",
		"-ts", "-tshort",
		"-iec",
	)
	parm, args := parms.New(args,
		"-l", "-loops",
		"-f", "-family",
		"-n", "-netns",
		"-rc", "-rcvbuf",
	)
	flag.Akas(
		flags.Aka{"-human", []string{"-human-readable"}},
		flags.Aka{"-s", []string{"-stats", "-statistics"}},
		flags.Aka{"-d", []string{"-details"}},
		flags.Aka{"-o", []string{"-oneline"}},
		flags.Aka{"-r", []string{"-resolve"}},
		flags.Aka{"-a", []string{"-all"}},
		flags.Aka{"-c", []string{"-color"}},
		flags.Aka{"-t", []string{"-timestamp"}},
		flags.Aka{"-ts", []string{"-tshort"}},
	)
	parm.Akas(
		parms.Aka{"-l", []string{"-loops"}},
		parms.Aka{"-f", []string{"-family"}},
		parms.Aka{"-n", []string{"-netns"}},
		parms.Aka{"-rc", []string{"-recbuf"}},
	)
	family, args := flags.New(args, "-4", "-6", "-B", "-M", "-0")
	switch {
	case family["-4"]:
		parm["-f"] = "inet"
	case family["-6"]:
		parm["-f"] = "inet6"
	case family["-B"]:
		parm["-f"] = "bridge"
	case family["-M"]:
		parm["-f"] = "mpls"
	case family["-0"]:
		parm["-f"] = "link"
	}
	return flag, parm, args
}
