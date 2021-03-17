// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package install

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/external/parms"
	"github.com/platinasystems/goes/lang"

	"github.com/satori/uuid"
)

type Command struct {
	g *goes.Goes

	AdminUser string
	AdminPass string

	Archive string

	DebootstrapOptions string

	Components string

	Daemons []string

	DebianDistro   string
	DebianDownload string

	DebootstrapProgram string

	DefaultArchive  string
	DefaultDNS      string
	DefaultHostname string
	DefaultRelease  string

	DNSAddr string

	GPGServer string

	Hostname string

	InstallDev string

	MgmtEth string
	MgmtIP  string
	MgmtGW  string

	PlatinaDistro   string
	PlatinaDownload string
	PlatinaGPG      string
	PlatinaRelease  string

	Target string

	UUIDEFI   uuid.UUID
	UUIDLinux uuid.UUID
	UUIDSwap  uuid.UUID
}

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (*Command) String() string { return "install" }

func (*Command) Usage() string {
	return "install OS"
}

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "install an operating system",
	}
}

func (c *Command) Man() lang.Alt {
	c.updateDNS()
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Install an operating system. The install command is used to
	install a new operating system, wiping out anything that may
	be presently installed.

	In the simplest case, typing "install" with no arguments will
	install the current Platina-verified Debian release.

USAGE
	install [options] [installer URL]

	[options] are listed below. The installer URL is generally not
	needed. It defaults to "` + c.DefaultArchive + `".

OPTIONS
	-admin-user USER	Sets the admin account. Default is platina

	-admin-pass PASS	Sets the admin password. Default is plat1na

	-allow-unauthenticated	Allow unauthenticated packages. Use with
				caution. Passed to debootstrap.

	-components COMP	Components to install. Passed to debootstrap

	-debian-distro DIST	Debian distro to install. Default is stretch

	-debian-download URL	Download URL. Default is
				http://ftp.debian.org/debian

	-debootstrap-program PRG Program to bootstrap Debian. Default is
				debootstrap

	-debug			Enable debugging. Passed to debootstrap

	-dnsaddr		Set default DNS addresses. Default is
				` + c.DefaultDNS + `

	-gpg-server srv		GPG server to validate keys. Default is
				pool.sks-keyservers.net

	-hostname name		Hostname to set. Default is ` +
			c.DefaultHostname + `

	-install-dev DEV	Device to install upon. Default is sda

	-mgmt-eth IF		Management ethernet. Default is enp5s0

	-mgmt-ip IP		Management IP address. Default is the currently
				configured management IP address

	-mgmt-GW GW		Management gateway IP address. Default is
				the currently configured default gateway.

	-platina-distro DIST	Platina distribution. Default is stretch

	-platina-download URL	Platina distribution download URL. Default
				is https://platina.io/goes/debian

	-platina-GPG key	Platina code signing key. Default is
				5FC2206DDF5ACEEA

	-platina-release NAME	Platina release. Default is ` +
			c.DefaultRelease + `

	-shell			Do not install. Just run a shell`,
	}
}

func (c *Command) Main(args ...string) error {
	c.updateDNS()

	parmTable := []struct {
		parm   string
		strPtr *string
		def    string
	}{
		{"-admin-user", &c.AdminUser, "platina"},
		{"-admin-pass", &c.AdminPass, "plat1na"},

		{"-components", &c.Components, ""},

		{"-debian-distro", &c.DebianDistro, "stretch"},
		{"-debian-download", &c.DebianDownload,
			"http://ftp.debian.org/debian"},

		{"-debootstrap-program", &c.DebootstrapProgram, "debootstrap"},

		{"-dnsaddr", &c.DNSAddr, c.DefaultDNS},

		{"-gpg-server", &c.GPGServer, "pool.sks-keyservers.net"},

		{"-hostname", &c.Hostname, c.DefaultHostname},

		{"-install-dev", &c.InstallDev, "sda"},

		{"-mgmt-eth", &c.MgmtEth, "enp5s0"},
		{"-mgmt-ip", &c.MgmtIP, ""},
		{"-mgmt-gw", &c.MgmtGW, ""},

		{"-platina-distro", &c.PlatinaDistro, "stretch"},
		{"-platina-download", &c.PlatinaDownload,
			"https://platina.io/goes/debian"},
		{"-platina-gpg", &c.PlatinaGPG, "5FC2206DDF5ACEEA"}, // kph@platinasystems.com key
		{"-platina-release", &c.PlatinaRelease, c.DefaultRelease},
	}

	parm, args := parms.New(args)
	flag, args := flags.New(args, "-shell", "-allow-unauthenticated",
		"-debug")

	for _, x := range parmTable {
		parm.ByName[x.parm] = ""
	}

	args = parm.Parse(args)
	c.Archive = c.DefaultArchive

	if len(args) >= 1 {
		c.Archive = args[0]
		args = args[1:]
	}
	if len(args) != 0 {
		return fmt.Errorf("Unexpected: %v", args)
	}

	for _, x := range parmTable {
		if val := parm.ByName[x.parm]; val != "" {
			*x.strPtr = val
		}
		if *x.strPtr == "" {
			*x.strPtr = x.def
		}
	}

	c.UUIDEFI = uuid.NewV4()
	c.UUIDLinux = uuid.NewV4()
	c.UUIDSwap = uuid.NewV4()

	mgmtDev := "eth0" // default
	if c.MgmtGW != "" {
		if ip := net.ParseIP(c.MgmtGW); ip == nil {
			return fmt.Errorf("Error parsing gateway IP %s",
				c.MgmtGW)
		}
	} else {
		gw, mgmt, err := defaultGateway()
		if err != nil {
			return fmt.Errorf("Error determining default gateway %w: try --mgmt-gw option",
				err)
		}
		mgmtDev = mgmt
		c.MgmtGW = gw.String()
	}

	if c.MgmtIP != "" {
		_, _, err := net.ParseCIDR(c.MgmtIP)
		if err != nil {
			return fmt.Errorf("Error parsing IP %s: %w",
				c.MgmtIP, err)
		}
	} else {
		ip, err := ipFromInterface(mgmtDev, false)
		if err != nil {
			return fmt.Errorf("Error finding management IP: %w, use the -mgmt-ip option",
				err)
		}
		if ip == nil {
			return fmt.Errorf("Set a management IP address or use the -mgmt-ip option")
		}
		c.MgmtIP = ip.String()
	}

	c.Target = "/var/run/goes/install-" + strconv.Itoa(os.Getppid())
	syscall.Unmount(c.Target, syscall.MNT_DETACH) // In case of stale mounts

	err := os.MkdirAll(c.Target, 0755)
	if err != nil {
		return fmt.Errorf("Unable to MkdirAll %s: %w", c.Target, err)
	}
	err = syscall.Mount("tmpfs", c.Target, "tmpfs", 0, "")
	if err != nil {
		return fmt.Errorf("Unable to create tmpfs: %w", err)
	}
	err = syscall.Mount("", c.Target, "", syscall.MS_SHARED, "")
	if err != nil {
		return fmt.Errorf("Unable to remount tmpfs shared: %w", err)
	}

	err = c.readArchive()
	if err != nil {
		return fmt.Errorf("Error reading archive: %w", err)
	}

	fmt.Printf("Target directory is %s\n", c.Target)

	for _, daemon := range c.Daemons {
		c.g.Main("daemons", "stop", daemon)
	}

	defer func() {
		for _, daemon := range c.Daemons {
			c.g.Main("daemons", "start", daemon)
		}
	}()

	time.Sleep(time.Second) // give daemons time to exit

	if flag.ByName["-shell"] {
		return c.g.Main("!", "-cd", "/", "-chroot", c.Target, "-m",
			"/bin/sh")
	}

	if flag.ByName["-allow-unauthenticated"] {
		c.DebootstrapOptions = "--allow-unauthenticated "
	}
	if flag.ByName["-debug"] {
		c.DebootstrapOptions += "--debug "
	}
	if c.Components != "" {
		c.DebootstrapOptions += "--components " + c.Components + " "
	}
	err = c.filesystemSetup()
	if err != nil {
		return err
	}

	err = c.networkSetup()
	if err != nil {
		return err
	}

	return c.debianInstall()
}
