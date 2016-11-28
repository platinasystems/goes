// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package parms

import "strings"

type Parm map[string]string

// New parses NAME[= ]VALUE parameters from the given arguments.
func New(args []string, parms ...string) (Parm, []string) {
	parm := make(Parm)
	for _, s := range parms {
		parm[s] = ""
	}
	for i := 0; i < len(args); {
		if _, found := parm[args[i]]; found {
			if i < len(args)-1 {
				parm[args[i]] = args[i+1]
				copy(args[i:], args[i+2:])
				args = args[:len(args)-2]
			} else {
				args = args[:len(args)-1]
			}
		} else if eq := strings.Index(args[i], "="); eq > 0 {
			name, value := args[i][:eq], args[i][eq+1:]
			if _, found := parm[name]; found {
				parm[name] = value
				if i < len(args)-1 {
					copy(args[i:], args[i+1:])
				}
				args = args[:len(args)-1]
			} else {
				i++
			}
		} else {
			i++
		}
	}
	return parm, args
}
