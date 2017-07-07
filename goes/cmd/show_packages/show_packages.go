// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show_packages

import (
	"fmt"
	"os"
	"strings"

	. "github.com/platinasystems/go"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "show-packages"
	Apropos = "print package repos info"
	Usage   = "show-packages [ -KEY ]..."
)

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	if len(args) == 0 {
		_, err := WriteTo(os.Stdout)
		return err
	}
	maps := []map[string]string{Package}
	if Packages != nil {
		maps = append(maps, Packages()...)
	}
	for _, m := range maps {
		if ip, found := m["importpath"]; found {
			for _, arg := range args {
				k := strings.TrimLeft(arg, "-")
				if val, found := m[k]; found {
					fmt.Print(ip, "[", k, "]: ", val, "\n")
				}
			}
		}
	}
	return nil
}
