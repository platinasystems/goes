// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"fmt"
	"strings"
)

func Usage(v Usager) string {
	return fmt.Sprint("usage:\t", strings.TrimSpace(v.Usage()))
}

type Usager interface {
	Usage() string
}

func (g *Goes) Usage() string {
	usage := g.USAGE
	if len(usage) == 0 {
		usage = `
	goes COMMAND [ ARGS ]...
	goes COMMAND -[-]HELPER [ ARGS ]...
	goes HELPER [ COMMAND ] [ ARGS ]...
	goes [ -d ] [ -x ] [[ -f ][ - | SCRIPT ]]

	HELPER := { apropos | complete | help | man | usage }`
	}
	return usage
}

func (g *Goes) usage(args ...string) error {
	var u Usager = g
	if len(args) > 0 {
		u = g.ByName[args[0]]
		if u == nil {
			return fmt.Errorf("%s: not found", args[0])
		}
	}
	fmt.Println(Usage(u))
	return nil
}
