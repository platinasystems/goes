// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package grub

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/cmd/cli"
	"github.com/platinasystems/go/goes/cmd/echo"
	"github.com/platinasystems/go/goes/cmd/elsecmd"
	"github.com/platinasystems/go/goes/cmd/falsecmd"
	"github.com/platinasystems/go/goes/cmd/ficmd"
	"github.com/platinasystems/go/goes/cmd/function"
	"github.com/platinasystems/go/goes/cmd/grub/background_color"
	"github.com/platinasystems/go/goes/cmd/grub/clear"
	"github.com/platinasystems/go/goes/cmd/grub/export"
	"github.com/platinasystems/go/goes/cmd/grub/gfxmode"
	"github.com/platinasystems/go/goes/cmd/grub/initrd"
	"github.com/platinasystems/go/goes/cmd/grub/insmod"
	"github.com/platinasystems/go/goes/cmd/grub/linux"
	"github.com/platinasystems/go/goes/cmd/grub/loadfont"
	"github.com/platinasystems/go/goes/cmd/grub/menuentry"
	"github.com/platinasystems/go/goes/cmd/grub/recordfail"
	"github.com/platinasystems/go/goes/cmd/grub/search"
	"github.com/platinasystems/go/goes/cmd/grub/set"
	"github.com/platinasystems/go/goes/cmd/grub/submenu"
	"github.com/platinasystems/go/goes/cmd/grub/terminal_output"

	"github.com/platinasystems/go/goes/cmd/ifcmd"
	"github.com/platinasystems/go/goes/cmd/kexec"
	"github.com/platinasystems/go/goes/cmd/testcmd"
	"github.com/platinasystems/go/goes/cmd/thencmd"
	"github.com/platinasystems/go/goes/cmd/truecmd"
	"github.com/platinasystems/go/goes/lang"
)

type Command struct {
}

var Goes = &goes.Goes{
	NAME: "grub",
	APROPOS: lang.Alt{
		lang.EnUS: "execute a grub configuration file",
	},
	ByName: map[string]cmd.Cmd{
		"background_color": background_color.Command{},
		"clear":            clear.Command{},
		"cli":              &cli.Command{},
		"echo":             echo.Command{},
		"else":             &elsecmd.Command{},
		"export":           export.Command{},
		"false":            falsecmd.Command{},
		"fi":               &ficmd.Command{},
		"function":         &function.Command{},
		"gfxmode":          gfxmode.Command{},
		"if":               &ifcmd.Command{},
		"initrd":           &initrd.Command{},
		"insmod":           insmod.Command{},
		"kexec":            kexec.Command{},
		"linux":            &linux.Command{},
		"loadfont":         loadfont.Command{},
		"menuentry":        &menuentry.Command{},
		"recordfail":       recordfail.Command{},
		"search":           search.Command{},
		"set":              &set.Command{},
		"submenu":          submenu.Command{},
		"[":                testcmd.Command{},
		"terminal_output":  terminal_output.Command{},
		"then":             &thencmd.Command{},
		"true":             truecmd.Command{},
	},
}

func (c *Command) Apropos() lang.Alt {
	return Goes.Apropos()
}

func (c *Command) Main(args ...string) error {
	return Goes.Main(args...)
}

func (c *Command) String() string {
	return Goes.String()
}

func (c *Command) Usage() string {
	return Goes.Usage()
}
