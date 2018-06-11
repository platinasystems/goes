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
	"regexp"
	"strconv"
	"strings"
	"time"

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
	"github.com/platinasystems/go/goes/cmd/platina/mk1/bootc"

	"github.com/platinasystems/go/goes/cmd/ifcmd"
	"github.com/platinasystems/go/goes/cmd/kexec"
	"github.com/platinasystems/go/goes/cmd/testcmd"
	"github.com/platinasystems/go/goes/cmd/thencmd"
	"github.com/platinasystems/go/goes/cmd/truecmd"
	"github.com/platinasystems/go/goes/lang"

	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/go/internal/url"

	"github.com/platinasystems/liner"
)

type Command struct{}

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
		"search":           &search.Command{},
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
	parm, args := parms.New(args, "-t")
	n := "/boot/grub/grub.cfg"
	if len(args) > 0 {
		n = args[0]
	}
	script, err := url.Open(n)
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

	err = Goes.Main()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Grub script returned %s\n", err)
	}

	root := Goes.EnvMap["root"]
	fmt.Printf("Root is %s translated %s\n", root, c.GetRoot())

	if kexec := bootc.Bootc(); len(kexec) > 1 {
		fmt.Printf("STRING: %v", kexec)
		err = Goes.Main(kexec...)
		fmt.Println(err)
	}

	c.ServeMenus() // FIXME so wrong

	menlen := len(Menuentry.Menus)
	if menlen == 0 && len(Linux.Kern) == 0 {
		fmt.Fprintf(os.Stderr, "Grub script did not define any menus or set a kernel\n")
	}

	if len(Linux.Kern) > 0 {
		kexec := c.KexecCommand()
		yn, err := c.readline(parm, fmt.Sprintf("Execute %s? <Yes/no> ", kexec), "Yes")
		if err != nil {
			return err
		}
		if strings.HasPrefix(yn, "Y") ||
			strings.HasPrefix(yn, "y") {
			err := Goes.Main(kexec...)
			return err
		}
		if err != nil {
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
	var menuItem int
	err = func() error {
		def := Goes.EnvMap["default"]
		if def == "" {
			def = "0"
		}
		mi, err := c.readline(parm, fmt.Sprintf("Menu item [%s]? ", def), def)
		if err != nil {
			return err
		}
		menuItem, err = strconv.Atoi(mi)
		if err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		return err
	}
	fmt.Printf("Running %d\n", menuItem)
	me := Menuentry.Menus[menuItem]
	fmt.Printf("Running menu item #%d:\n", menuItem)
	err = me.RunFun(os.Stdin, os.Stdout, os.Stderr, false, false)
	fmt.Printf("Kernel defined: %s\n", Linux.Kern)
	fmt.Printf("Linux command: %v\n", Linux.Cmd)
	fmt.Printf("Initrd: %v\n", Initrd.Initrd)

	root = Goes.EnvMap["root"]
	fmt.Printf("Root is %s translated %s\n", root, c.GetRoot())

	if len(Linux.Kern) > 0 {
		kexec := c.KexecCommand()
		yn, err := c.readline(parm, fmt.Sprintf("Execute %s? <Yes/no> ", kexec), "Yes")
		if err != nil {
			return err
		}
		if strings.HasPrefix(yn, "Y") ||
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

func (c *Command) GetRoot() string {
	root := Goes.EnvMap["root"]
	if root == "" {
		return root
	}

	dev := root
	if !strings.HasPrefix(root, "/dev/") {
		re := regexp.MustCompile(`^((hd(?P<Unit>\d+)),.*(?P<Partition>\d+))$`)
		r := re.FindStringSubmatch(root)
		if len(r) == 5 {
			unit, err := strconv.Atoi(r[3])
			if err == nil {
				dev = "/dev/sd" + string(97+unit) + r[4]
			}
		}
	}
	trans, err := c.findMountedFS(dev)
	if err != nil {
		panic(err)
	}
	if trans != "" {
		if trans != "/" {
			return trans
		}
		return ""
	}
	return dev
}

func (c *Command) KexecCommand() []string {
	k := Linux.Kern
	i := Initrd.Initrd
	if k[0] != '/' {
		k = "/" + k
	}
	if i[0] != '/' {
		i = "/" + i
	}
	k = c.GetRoot() + k
	i = c.GetRoot() + i
	return []string{"kexec", "-k", k, "-i", i, "-c", strings.Join(Linux.Cmd, " "), "-e"}

}

func (c *Command) readline(parm *parms.Parms, prompt string, def string) (string, error) {
	line := liner.NewLiner()
	defer line.Close()
	if parm.ByName["-t"] != "" {
		timeout, err := time.ParseDuration(parm.ByName["-t"])
		if err != nil {
			return "", err
		}
		err = line.SetDuration(timeout)
		if err != nil {
			return "", err
		}
	}

	line.SetCtrlCAborts(true)

	mi, err := line.Prompt(prompt)
	if err != nil {
		if err == liner.ErrTimeOut {
			mi = ""
			fmt.Println("<timeout>")
		} else {
			return "", err
		}
	}

	if mi == "" {
		mi = def
	}
	return mi, nil
}

func (c *Command) findMountedFS(fs string) (string, error) {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if fs == fields[0] {
			return fields[1], nil
		}
	}
	return "", scanner.Err()

}
