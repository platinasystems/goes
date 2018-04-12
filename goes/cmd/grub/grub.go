// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package grub

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

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
	"github.com/platinasystems/go/goes/cmd/grub/webserver"

	"github.com/platinasystems/go/goes/cmd/ifcmd"
	"github.com/platinasystems/go/goes/cmd/kexec"
	"github.com/platinasystems/go/goes/cmd/testcmd"
	"github.com/platinasystems/go/goes/cmd/thencmd"
	"github.com/platinasystems/go/goes/cmd/truecmd"
	"github.com/platinasystems/go/goes/lang"

	"github.com/platinasystems/go/internal/url"
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
		"initrd":           Initrd,
		"insmod":           insmod.Command{},
		"kexec":            kexec.Command{},
		"linux":            Linux,
		"loadfont":         loadfont.Command{},
		"menuentry":        Menuentry,
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

var Linux = &linux.Command{}

var Initrd = &initrd.Command{}

var Menuentry = &menuentry.Command{}

func (c *Command) Apropos() lang.Alt {
	return Goes.Apropos()
}

func (c *Command) Main(args ...string) error {
	script, err := url.Open("/boot/grub/grub.cfg")
	if err != nil {
		return err
	}
	defer script.Close()

	scanner := bufio.NewScanner(script)

	Goes.Catline = func(prompt string) (string, error) {
		if scanner.Scan() {
			return scanner.Text(), nil
		}
		err := scanner.Err()
		if err == nil {
			err = io.EOF
		}
		return "", err
	}

	err = Goes.Main(args...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Grub script returned %s\n", err)
	}

	webserver.ServeMenus(Goes) // FIXME so wrong

	menlen := len(Menuentry.Menus)
	if menlen == 0 && len(Linux.Kern) == 0 {
		fmt.Fprintf(os.Stderr, "Grub script did not define any menus or set a kernel\n")
	}

	if len(Linux.Kern) > 0 {
		kexec := []string{"kexec", "-k", Linux.Kern, "-i", Initrd.Initrd, "-c", strings.Join(Linux.Cmd, " "), "-e"}
		fmt.Printf("Execute %s? <Yes/no> ", kexec)
		yn := ""
		_, err := fmt.Fscanln(os.Stdin, &yn)
		if err != nil {
			return err
		}
		if yn == "" || strings.HasPrefix(yn, "Y") ||
			strings.HasPrefix(yn, "y") {
			err := Goes.Main(kexec...)
			return err
		}
	}

	if menlen == 0 {
		return errors.New("No defined kernel or menus")
	}
	fmt.Printf("Menus defined: %d\n", menlen)
	for i, me := range Menuentry.Menus {
		fmt.Printf("[%d]   %s\n", i, me.Name)
	}
	fmt.Printf("Menu item [%d]? ", 0) //FIXME get the real default
	mi := ""                          // FIXME get the real default
	_, err = fmt.Fscanln(os.Stdin, &mi)
	if err != nil {
		return err
	}

	menuItem, err := strconv.Atoi(mi)
	fmt.Printf("Running %d\n", menuItem)
	me := Menuentry.Menus[menuItem]
	fmt.Printf("Running menu item #%d:\n", menuItem)
	err = me.RunFun(os.Stdin, os.Stdout, os.Stderr, false, false)
	fmt.Printf("Kernel defined: %s\n", Linux.Kern)
	fmt.Printf("Linux command: %v\n", Linux.Cmd)
	fmt.Printf("Initrd: %v\n", Initrd.Initrd)

	if len(Linux.Kern) > 0 {
		kexec := []string{"kexec", "-k", Linux.Kern, "-i", Initrd.Initrd, "-c", strings.Join(Linux.Cmd, " "), "-e"}
		fmt.Printf("Execute %s? <Yes/no> ", kexec)
		yn := ""
		_, err := fmt.Fscanln(os.Stdin, &yn)
		if err != nil {
			return err
		}
		if yn == "" || strings.HasPrefix(yn, "Y") ||
			strings.HasPrefix(yn, "y") {
			err := Goes.Main(kexec...)
			return err
		}
	}

	return err
}

func (c *Command) String() string {
	return Goes.String()
}

func (c *Command) Usage() string {
	return Goes.Usage()
}
