// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"time"

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
	"github.com/platinasystems/go/goes/cmd/dmesg"
	"github.com/platinasystems/go/goes/cmd/echo"
	eepromcmd "github.com/platinasystems/go/goes/cmd/eeprom"
	eeprom "github.com/platinasystems/go/goes/cmd/eeprom/platina_eeprom"
	"github.com/platinasystems/go/goes/cmd/elsecmd"
	"github.com/platinasystems/go/goes/cmd/env"
	"github.com/platinasystems/go/goes/cmd/exec"
	"github.com/platinasystems/go/goes/cmd/exit"
	"github.com/platinasystems/go/goes/cmd/export"
	"github.com/platinasystems/go/goes/cmd/falsecmd"
	"github.com/platinasystems/go/goes/cmd/femtocom"
	"github.com/platinasystems/go/goes/cmd/ficmd"
	"github.com/platinasystems/go/goes/cmd/function"
	"github.com/platinasystems/go/goes/cmd/gpio"
	"github.com/platinasystems/go/goes/cmd/hdel"
	"github.com/platinasystems/go/goes/cmd/hdelta"
	"github.com/platinasystems/go/goes/cmd/hexists"
	"github.com/platinasystems/go/goes/cmd/hget"
	"github.com/platinasystems/go/goes/cmd/hgetall"
	"github.com/platinasystems/go/goes/cmd/hkeys"
	"github.com/platinasystems/go/goes/cmd/hset"
	"github.com/platinasystems/go/goes/cmd/i2c"
	"github.com/platinasystems/go/goes/cmd/i2cd"
	"github.com/platinasystems/go/goes/cmd/ifcmd"
	"github.com/platinasystems/go/goes/cmd/iminfo"
	"github.com/platinasystems/go/goes/cmd/imx6d"
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
	"github.com/platinasystems/go/goes/cmd/platina/mk2/lc1/bmc/diag"
	"github.com/platinasystems/go/goes/cmd/platina/mk2/lc1/bmc/upgrade"
	"github.com/platinasystems/go/goes/cmd/platina/mk2/lc1/bmc/upgraded"
	"github.com/platinasystems/go/goes/cmd/ps"
	"github.com/platinasystems/go/goes/cmd/pwd"
	"github.com/platinasystems/go/goes/cmd/reboot"
	"github.com/platinasystems/go/goes/cmd/redisd"
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
	"github.com/platinasystems/go/goes/cmd/subscribe"
	"github.com/platinasystems/go/goes/cmd/sync"
	"github.com/platinasystems/go/goes/cmd/testcmd"
	"github.com/platinasystems/go/goes/cmd/thencmd"
	"github.com/platinasystems/go/goes/cmd/truecmd"
	"github.com/platinasystems/go/goes/cmd/umount"
	"github.com/platinasystems/go/goes/cmd/uninstall"
	"github.com/platinasystems/go/goes/cmd/uptimed"
	"github.com/platinasystems/go/goes/cmd/watchdog"
	"github.com/platinasystems/go/goes/cmd/wget"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/redis/publisher"
)

var Goes = &goes.Goes{
	NAME: "goes-" + name,
	APROPOS: lang.Alt{
		lang.EnUS: "platina's mk2 lc1 baseboard management controller",
	},
	ByName: map[string]cmd.Cmd{
		"!":     bang.Command{},
		"boot":  &boot.Command{},
		"cat":   cat.Command{},
		"cd":    &cd.Command{},
		"chmod": chmod.Command{},
		"cli":   &cli.Command{},
		"cp":    cp.Command{},
		"diag": &diag.Command{
			Gpio: gpioInit,
		},
		"dmesg":  dmesg.Command{},
		"echo":   echo.Command{},
		"eeprom": eepromcmd.Command{},
		"else":   &elsecmd.Command{},
		"env":    &env.Command{},
		"exec":   exec.Command{},
		"exit":   exit.Command{},
		"export": export.Command{},
		"false":  falsecmd.Command{},
		/*FIXME
		"fantrayd": &fantrayd.Command{
			Init: fantraydInit,
		},
		*/
		"femtocom": femtocom.Command{},
		"fi":       &ficmd.Command{},
		/*FIXME
		"fspd": &fspd.Command{
			Init: fspdInit,
			Gpio: gpioInit,
		},
		*/
		"function": &function.Command{},
		"goes-daemons": &daemons.Command{
			Init: [][]string{
				[]string{"redisd"},
				//[]string{"fantrayd"},
				//[]string{"fspd"},
				[]string{"i2cd"},
				[]string{"imx6d"},
				//[]string{"ledgpiod"},
				[]string{"upgraded"},
				[]string{"uptimed"},
				//[]string{"ucd9090d"},
				//]string{"w83795d"},
			},
		},
		"gpio": &gpio.Command{
			Init: gpioInit,
		},
		"hdel":    hdel.Command{},
		"hdelta":  &hdelta.Command{},
		"hexists": hexists.Command{},
		"hget":    hget.Command{},
		"hgetall": hgetall.Command{},
		"hkeys":   hkeys.Command{},
		"hset":    hset.Command{},
		"i2c":     i2c.Command{},
		"i2cd": &i2cd.Command{
			Gpio: gpioInit,
		},
		"if": &ifcmd.Command{},
		"imx6d": &imx6d.Command{
			VpageByKey: map[string]uint8{
				"bmc.temperature.units.C": 1,
			},
		},
		"insmod":  insmod.Command{},
		"install": &install.Command{},
		"ip":      ip.Goes,
		"kexec":   kexec.Command{},
		"keys":    keys.Command{},
		"kill":    kill.Command{},
		/*FIXME
		"ledgpiod": &ledgpiod.Command{
			Init: ledgpiodInit,
		},
		*/
		/*FIXME
		"lcabsd": lcabsd.Command{},
		*/
		"ln":    ln.Command{},
		"log":   log.Command{},
		"ls":    ls.Command{},
		"lsmod": lsmod.Command{},
		"mkdir": mkdir.Command{},
		"mknod": mknod.Command{},
		"mount": mount.Command{},
		"ping":  ping.Command{},
		"ps":    ps.Command{},
		"pwd":   pwd.Command{},
		/*FIXME
		"qsfpd": qsfpd.Command{},
		*/
		"reboot": reboot.Command{},
		"redisd": &redisd.Command{
			Devs:    []string{"lo", "eth0"},
			Machine: "platina-mk2-lc1-bmc",
			Hook: func(pub *publisher.Publisher) {
				eeprom.Config(
					eeprom.BusIndex(0),
					eeprom.BusAddress(0x55),
					eeprom.BusDelay(10*time.Millisecond),
					eeprom.MinMacs(2),
					eeprom.OUI([3]byte{0x02, 0x46, 0x8a}),
				)
			},
		},
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
		"/init":  &slashinit.Command{},
		"sleep":  sleep.Command{},
		"source": &source.Command{},
		"start": &start.Command{
			ConfGpioHook: startConfGpioHook,
		},
		"stop":      &stop.Command{},
		"stty":      stty.Command{},
		"subscribe": subscribe.Command{},
		"sync":      sync.Command{},
		"[":         testcmd.Command{},
		"then":      &thencmd.Command{},
		/*FIXME
		"toggle": &toggle.Command{},
		*/
		"true": truecmd.Command{},
		/*FIXME
		"ucd9090d": &ucd9090d.Command{
			Init: ucd9090dInit,
			Gpio: gpioInit,
		},
		*/
		"umount":    umount.Command{},
		"uninstall": &uninstall.Command{},
		"upgrade": &upgrade.Command{
			Gpio: gpioInit,
		},
		"upgraded": &upgraded.Command{},
		"uptimed":  uptimed.Command(make(chan struct{})),
		/*FIXME
		"w83795d": &w83795d.Command{
			Init: w83795dInit,
		},
		*/
		"watchdog": &watchdog.Command{
			GpioPin: "BMC_WDI",
			Init:    gpioInit,
		},
		"wget": wget.Command{},
	},
}
