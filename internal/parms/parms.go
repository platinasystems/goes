// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package parms

import (
	"errors"
	"strings"
)

var ErrNotFound = errors.New("not found")

type Parm map[string]string

// Parses {NAME VALUE} and NAME=VALUE parameters from the given arguments.
func New(args []string, parms ...string) (Parm, []string) {
	parm := make(Parm)
	for _, s := range parms {
		parm[s] = ""
	}
	for i := 0; i < len(args); {
		if eq := strings.Index(args[i], "="); eq > 0 {
			if parm.Set(args[i][:eq], args[i][eq+1:]) == nil {
				if i < len(args)-1 {
					copy(args[i:], args[i+1:])
				}
				args = args[:len(args)-1]
			} else {
				i++
			}
		} else if i < len(args)-1 &&
			parm.Set(args[i], args[i+1]) == nil {
			copy(args[i:], args[i+2:])
			args = args[:len(args)-2]
			i += 2
		} else {
			i++
		}
	}
	return parm, args
}

// Set will concatenate a non empty parm
func (parm Parm) Set(name, value string) error {
	cur, found := parm[name]
	if !found {
		return ErrNotFound
	}
	if len(cur) > 0 && len(value) > 0 {
		parm[name] = cur + " " + value
	} else {
		parm[name] = value
	}
	return nil
}
