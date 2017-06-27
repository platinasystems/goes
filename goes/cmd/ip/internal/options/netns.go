// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import (
	"github.com/platinasystems/go/goes/cmd/ip/internal/netns"
	"github.com/platinasystems/go/internal/parms"
)

// Parse [{-n | -netns} NETNS] and switch to the given namespace
func Netns(args []string) ([]string, error) {
	var err error
	parm, args := parms.New(args, "-n", "-netns")
	parm.Akas(parms.Aka{"-n", []string{"-netns"}})
	if name := parm["-n"]; len(name) > 0 {
		err = netns.Switch(name)
	}
	return args, err

}
