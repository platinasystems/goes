// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import "fmt"

type helper interface {
	Help(...string) string
}

func (g *Goes) Help(args ...string) string {
	g.swap(args)
	g.shift(args)
	if len(args) > 0 {
		if v, found := g.ByName[args[0]]; found {
			if method, found := v.(helper); found {
				return method.Help(args[1:]...)
			}
			return Usage(v)
		}
	}
	return Usage(g)
}

func (g *Goes) help(args ...string) error {
	h := g.Help(args...)
	if len(h) > 0 {
		fmt.Println(h)
	}
	return nil
}
