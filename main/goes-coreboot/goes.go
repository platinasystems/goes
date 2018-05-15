// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/cmd/bang"
	"github.com/platinasystems/go/goes/cmd/boot"
	"github.com/platinasystems/go/goes/cmd/cat"
	"github.com/platinasystems/go/goes/cmd/cd"
	"github.com/platinasystems/go/goes/cmd/chmod"
	"github.com/platinasystems/go/goes/cmd/cli"
	"github.com/platinasystems/go/goes/cmd/cmdline"
	"github.com/platinasystems/go/goes/cmd/cp"
	"github.com/platinasystems/go/goes/cmd/daemons"
	"github.com/platinasystems/go/goes/cmd/dhcpcd"
	"github.com/platinasystems/go/goes/cmd/dmesg"
	"github.com/platinasystems/go/goes/cmd/echo"
	"github.com/platinasystems/go/goes/cmd/elsecmd"
	"github.com/platinasystems/go/goes/cmd/env"
	"github.com/platinasystems/go/goes/cmd/exec"
	"github.com/platinasystems/go/goes/cmd/exit"
	"github.com/platinasystems/go/goes/cmd/export"
	"github.com/platinasystems/go/goes/cmd/falsecmd"
	"github.com/platinasystems/go/goes/cmd/femtocom"
	"github.com/platinasystems/go/goes/cmd/ficmd"
	"github.com/platinasystems/go/goes/cmd/function"
	"github.com/platinasystems/go/goes/cmd/grub"
	"github.com/platinasystems/go/goes/cmd/ifcmd"
	"github.com/platinasystems/go/goes/cmd/iminfo"
	"github.com/platinasystems/go/goes/cmd/insmod"
	"github.com/platinasystems/go/goes/cmd/install"
	"github.com/platinasystems/go/goes/cmd/ip"
	"github.com/platinasystems/go/goes/cmd/kexec"
	"github.com/platinasystems/go/goes/cmd/keys"
	"github.com/platinasystems/go/goes/cmd/kill"
	"github.com/platinasystems/go/goes/cmd/ln"
	"github.com/platinasystems/go/goes/cmd/log"
	"github.com/platinasystems/go/goes/cmd/ls"
	"github.com/platinasystems/go/goes/cmd/lsmod"
	"github.com/platinasystems/go/goes/cmd/mkdir"
	"github.com/platinasystems/go/goes/cmd/mknod"
	"github.com/platinasystems/go/goes/cmd/mount"
	"github.com/platinasystems/go/goes/cmd/ping"
	"github.com/platinasystems/go/goes/cmd/platina/mk1/bootc"
	"github.com/platinasystems/go/goes/cmd/ps"
	"github.com/platinasystems/go/goes/cmd/pwd"
	"github.com/platinasystems/go/goes/cmd/reboot"
	"github.com/platinasystems/go/goes/cmd/reload"
	"github.com/platinasystems/go/goes/cmd/restart"
	"github.com/platinasystems/go/goes/cmd/rm"
	"github.com/platinasystems/go/goes/cmd/rmmod"
	"github.com/platinasystems/go/goes/cmd/slashinit"
	"github.com/platinasystems/go/goes/cmd/sleep"
	"github.com/platinasystems/go/goes/cmd/source"
	"github.com/platinasystems/go/goes/cmd/start"
	"github.com/platinasystems/go/goes/cmd/stop"
	"github.com/platinasystems/go/goes/cmd/stty"
	"github.com/platinasystems/go/goes/cmd/sync"
	"github.com/platinasystems/go/goes/cmd/testcmd"
	"github.com/platinasystems/go/goes/cmd/thencmd"
	"github.com/platinasystems/go/goes/cmd/truecmd"
	"github.com/platinasystems/go/goes/cmd/umount"
	"github.com/platinasystems/go/goes/cmd/uninstall"
	"github.com/platinasystems/go/goes/cmd/uptimed"
	"github.com/platinasystems/go/goes/cmd/wget"
	"github.com/platinasystems/go/goes/lang"
)

var Goes = &goes.Goes{
	NAME: "goes-" + name,
	APROPOS: lang.Alt{
		lang.EnUS: "the coreboot goes machine",
	},
	ByName: map[string]cmd.Cmd{
		"!":        bang.Command{},
		"cli":      &cli.Command{},
		"boot":     &boot.Command{},
		"bootc":    &bootc.Command{},
		"cat":      cat.Command{},
		"cd":       &cd.Command{},
		"chmod":    chmod.Command{},
		"cp":       cp.Command{},
		"dhcpcd":   &dhcpcd.Command{},
		"dmesg":    dmesg.Command{},
		"echo":     echo.Command{},
		"else":     &elsecmd.Command{},
		"env":      &env.Command{},
		"exec":     exec.Command{},
		"exit":     exit.Command{},
		"export":   export.Command{},
		"false":    falsecmd.Command{},
		"femtocom": femtocom.Command{},
		"fi":       &ficmd.Command{},
		"function": &function.Command{},
		"goes-daemons": &daemons.Command{
			Init: [][]string{
				[]string{"dhcpcd"},
			},
		},
		"grub":    &grub.Command{},
		"if":      &ifcmd.Command{},
		"insmod":  insmod.Command{},
		"install": &install.Command{},
		"ip":      ip.Goes,
		"kexec":   kexec.Command{},
		"keys":    keys.Command{},
		"kill":    kill.Command{},
		"ln":      ln.Command{},
		"log":     log.Command{},
		"ls":      ls.Command{},
		"lsmod":   lsmod.Command{},
		"mkdir":   mkdir.Command{},
		"mknod":   mknod.Command{},
		"mount":   mount.Command{},
		"ping":    ping.Command{},
		"ps":      ps.Command{},
		"pwd":     pwd.Command{},
		"reboot":  reboot.Command{},

		"reload":  reload.Command{},
		"restart": &restart.Command{},
		"rm":      rm.Command{},
		"rmmod":   rmmod.Command{},

		"show": &goes.Goes{
			NAME:  "show",
			USAGE: "show OBJECT",
			APROPOS: lang.Alt{
				lang.EnUS: "print stuff",
			},
			ByName: map[string]cmd.Cmd{
				"cmdline":  cmdline.Command{},
				"iminfo":   iminfo.Command{},
				"machine":  goes.ShowMachine(name),
				"packages": goes.ShowPackages{},
			},
		},

		"/init":     &slashinit.Command{},
		"sleep":     sleep.Command{},
		"source":    &source.Command{},
		"start":     &start.Command{},
		"stop":      &stop.Command{},
		"stty":      stty.Command{},
		"sync":      sync.Command{},
		"[":         testcmd.Command{},
		"then":      &thencmd.Command{},
		"true":      truecmd.Command{},
		"umount":    umount.Command{},
		"uninstall": &uninstall.Command{},
		"uptimed":   uptimed.Command(make(chan struct{})),
		"wget":      wget.Command{},
	},
}
