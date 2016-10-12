// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package export

import (
	"fmt"
	"os"
	"strings"
)

type export struct{}

func New() export { return export{} }

func (export) String() string { return "export" }
func (export) Tag() string    { return "builtin" }
func (export) Usage() string  { return "export [NAME[=VALUE]]..." }

func (export) Main(args ...string) error {
	if len(args) == 0 {
		for _, nv := range os.Environ() {
			fmt.Println(nv)
		}
		return nil
	}
	for _, arg := range args {
		eq := strings.Index(arg, "=")
		if eq < 0 {
			if err := os.Unsetenv(arg); err != nil {
				return err
			}
		} else if err := os.Setenv(arg[:eq], arg[eq+1:]); err != nil {
			return err
		}
	}
	return nil
}

func (export) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "set process configuration",
	}
}

func (export) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	export - process configuration

SYNOPSIS
	export [NAME=[VALUE]]...

DESCRIPTION
	Configure the named process environment parameter.

	If no VALUE is given, NAME is reset.

	If no NAMES are supplied, a list of names of all exported variables
	is printed.`,
	}
}
