// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package iminfo

import (
	"fmt"
	"io/ioutil"

	"github.com/platinasystems/go/internal/fit"
	"github.com/platinasystems/go/internal/goes/lang"
)

const (
	Name    = "iminfo"
	Apropos = "FIXME"
	Usage   = "iminfo"
)

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

func (cmd) Main(args ...string) error {
	if n := len(args); n == 0 {
		return fmt.Errorf("DESTINATION: missing")
	} else if n > 1 {
		return fmt.Errorf("%v: unexpected", args[1:])
	}
	itb := args[0]
	b, err := ioutil.ReadFile(itb)
	if err != nil {
		return err
	}

	fit := fit.Parse(b)

	fmt.Printf("Description = %s\nAddressCells = %d\nTimeStamp = %s\n", fit.Description, fit.AddressCells, fit.TimeStamp)

	for name, cfg := range fit.Configs {
		fmt.Printf("Configuration %s:%s\n", name, (*cfg).Description)
		listImages(cfg.ImageList)
	}
	return nil
}

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name }

func listImages(imageList []*fit.Image) {
	for _, image := range imageList {
		fmt.Printf(`  %s:
    Description=%s
    Type=%s
    Arch=%s
    OS=%s
    Compression=%s
    LoadAddr=%x
`,
			image.Name, image.Description, image.Type,
			image.Arch, image.Os, image.Compression,
			image.LoadAddr)
	}
}

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
