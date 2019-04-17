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

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/cmd/cli"
	"github.com/platinasystems/goes/cmd/echo"
	"github.com/platinasystems/goes/cmd/elsecmd"
	"github.com/platinasystems/goes/cmd/falsecmd"
	"github.com/platinasystems/goes/cmd/ficmd"
	"github.com/platinasystems/goes/cmd/function"
	"github.com/platinasystems/goes/cmd/grub/background_color"
	"github.com/platinasystems/goes/cmd/grub/clear"
	"github.com/platinasystems/goes/cmd/grub/export"
	"github.com/platinasystems/goes/cmd/grub/gfxmode"
	"github.com/platinasystems/goes/cmd/grub/initrd"
	"github.com/platinasystems/goes/cmd/grub/insmod"
	"github.com/platinasystems/goes/cmd/grub/linux"
	"github.com/platinasystems/goes/cmd/grub/loadfont"
	"github.com/platinasystems/goes/cmd/grub/menuentry"
	"github.com/platinasystems/goes/cmd/grub/recordfail"
	"github.com/platinasystems/goes/cmd/grub/search"
	"github.com/platinasystems/goes/cmd/grub/set"
	"github.com/platinasystems/goes/cmd/grub/submenu"
	"github.com/platinasystems/goes/cmd/grub/terminal_output"

	"github.com/platinasystems/goes/cmd/ifcmd"
	"github.com/platinasystems/goes/cmd/kexec"
	"github.com/platinasystems/goes/cmd/testcmd"
	"github.com/platinasystems/goes/cmd/thencmd"
	"github.com/platinasystems/goes/cmd/truecmd"
	"github.com/platinasystems/goes/lang"

	"github.com/platinasystems/flags"
	"github.com/platinasystems/parms"
	"github.com/platinasystems/url"

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
		"submenu":          submenu.Command{M: Menuentry},
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
	flag, args := flags.New(args, "--daemon")

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

	m := Menuentry.Menus
	c.ServeMenus() // FIXME so wrong

	menlen := len(m)
	if menlen == 0 && len(Linux.Kern) == 0 {
		fmt.Fprintf(os.Stderr, "Grub script did not define any menus or set a kernel\n")
	}

	err = c.AskKernel(parm, flag)
	if err != nil {
		return err
	}
	if menlen == 0 {
		return errors.New("No defined kernel or menus")
	}
	fmt.Printf("Menus defined: %d\n", menlen)
	err = c.RunMenu(m, parm, flag)
	if err != nil {
		return err
	}
	root = Goes.EnvMap["root"]
	fmt.Printf("Root is %s translated %s\n", root, c.GetRoot())

	err = c.AskKernel(parm, flag)

	return err
}

func (c *Command) String() string {
	return Goes.String()
}

func (c *Command) Usage() string {
	return Goes.Usage()
}

func (c *Command) RunMenu(m []menuentry.Entry, parm *parms.Parms, flag *flags.Flags) (err error) {
	for len(m) != 0 {
		for i, me := range m {
			fmt.Printf("[%d]   %s\n", i, me.Name)
		}
		var menuItem int
		err = func() error {
			def := Goes.EnvMap["default"]
			if def == "" {
				def = "0"
			}
			mi, err := c.readline(parm, flag, fmt.Sprintf("Menu item [%s]? ", def), def)
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
			return
		}
		if menuItem >= len(m) {
			return errors.New("Menu item out of range")
		}
		me := m[menuItem]
		Menuentry.Menus = Menuentry.Menus[:0]
		err = me.RunFun(os.Stdin, os.Stdout, os.Stderr, false, false)
		fmt.Printf("Kernel defined: %s\n", Linux.Kern)
		fmt.Printf("Linux command: %v\n", Linux.Cmd)
		fmt.Printf("Initrd: %v\n", Initrd.Initrd)
		err = c.AskKernel(parm, flag)
		if err != nil {
			return err
		}
		m = Menuentry.Menus
	}
	return
}

func (c *Command) AskKernel(parm *parms.Parms, flag *flags.Flags) (err error) {
	if len(Linux.Kern) > 0 {
		kexec := c.KexecCommand()
		yn, err := c.readline(parm, flag, fmt.Sprintf("Execute %s? <Yes/no> ", kexec), "Yes")
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
	return
}

func (c *Command) GetRoot() string {
	root := Goes.EnvMap["root"]
	if root == "" {
		return root
	}

	devSD := root
	devHD := ""
	devVD := ""
	if !strings.HasPrefix(root, "/dev/") {
		re := regexp.MustCompile(`^((hd(?P<Unit>\d+)),.*(?P<Partition>\d+))$`)
		r := re.FindStringSubmatch(root)
		if len(r) == 5 {
			unit, err := strconv.Atoi(r[3])
			if err == nil {
				devSD = "/dev/sd" + string(97+unit) + r[4]
				devHD = "/dev/hd" + string(97+unit) + r[4]
				devVD = "/dev/vd" + string(97+unit) + r[4]
			}
		}
	}
	trans, err := c.findMountedFS(devSD)
	if err != nil && devHD != "" {
		trans, err = c.findMountedFS(devHD)
		if err != nil && devVD != "" {
			trans, err = c.findMountedFS(devVD)
		}
	}
	if trans != "" {
		if trans != "/" {
			return trans
		}
		return ""
	}
	return devSD
}

func (c *Command) KexecCommand() []string {
	k := Linux.Kern
	i := Initrd.Initrd
	if len(k) == 0 {
		return []string{}
	}
	if k[0] != '/' {
		k = "/" + k
	}
	if i[0] != '/' {
		i = "/" + i
	}
	k = c.GetRoot() + k
	i = c.GetRoot() + i
	co := false
	for _, cmd := range Linux.Cmd {
		if strings.HasPrefix(cmd, "console=") {
			co = true
			break
		}
	}
	cl := strings.TrimRight(strings.Join(Linux.Cmd, " "), " ")
	if !co {
		if cl != "" {
			cl = cl + " "
		}
		cl = cl + "console=ttyS0,115200n8"
	}
	return []string{"kexec", "-k", k, "-i", i, "-c", cl, "-e"}
}

func (c *Command) readline(parm *parms.Parms, flag *flags.Flags, prompt string, def string) (mi string, err error) {
	var timeout time.Duration
	tmEnv := Goes.EnvMap["timeout"]
	if tmEnv != "" {
		tm, err := strconv.Atoi(tmEnv)
		if err == nil {
			timeout = time.Duration(tm) * time.Second
		}
	}
	if timeout == 0 {
		if parm.ByName["-t"] != "" {
			timeout, err = time.ParseDuration(parm.ByName["-t"])
			if err != nil {
				return "", err
			}
		}
	}

	if flag.ByName["--daemon"] == false {
		line := liner.NewLiner()
		defer line.Close()
		line.SetCtrlCAborts(true)
		if timeout != 0 {
			err := line.SetDuration(timeout)
			if err != nil {
				return "", err
			}
		}

		mi, err = line.Prompt(prompt)
		if err != nil {
			if err == liner.ErrTimeOut {
				mi = ""
				fmt.Println("<timeout>")
			} else {
				return "", err
			}
		}
	} else {
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		time.Sleep(timeout)
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
