// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package license

import (
	"fmt"
	"sync"

	"github.com/platinasystems/go/copyright"
	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/goes/lang"
)

const (
	Name    = "license"
	Apropos = "print machine license(s)"
	Usage   = "license"
)

var (
	Init = func() {}
	once sync.Once
	// Some machines may have additional licenses.
	Others []Other
)

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type Other struct {
	Name, Text string
}

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

func (cmd) Kind() goes.Kind { return goes.DontFork }

func (cmd) Main(args ...string) error {
	once.Do(Init)
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

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

func prettyprint(title, text string) {
	fmt.Println(title)
	for _ = range title {
		fmt.Print("=")
	}
	fmt.Print("\n\n", text, "\n")
}

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
