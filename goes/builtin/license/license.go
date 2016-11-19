// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package license

import (
	"fmt"

	"github.com/platinasystems/go/copyright"
)

const Name = "license"

// Some machines may have additional licenses.
var Others []Other

type Other struct {
	Name, Text string
}

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Tag() string    { return "builtin" }
func (cmd) Usage() string  { return Name }

func (cmd) Main(args ...string) error {
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}
	prettyprint("github.com/platinasystems/go", copyright.License)
	for _, l := range Others {
		fmt.Print("\n\n")
		prettyprint(l.Name, l.Text)
	}
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print machine license(s)",
	}
}

func prettyprint(title, text string) {
	fmt.Println(title)
	for _ = range title {
		fmt.Print("=")
	}
	fmt.Print("\n\n", text, "\n")
}
