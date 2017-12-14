// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ping

import (
	"fmt"
	"net"
	"syscall"
	"time"

	"github.com/platinasystems/go/goes/lang"
	"github.com/tatsushid/go-fastping"
)

const (
	Name    = "ping"
	Apropos = "send ICMP ECHO_REQUEST to network host"
	Usage   = "ping DESTINATION"
	Man     = `
DESCRIPTION
	Send ICMP ECHO_REQUEST to given host and print ECHO_REPLY.`
)

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	Man() lang.Alt
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
	dest := args[0]
	pinger := fastping.NewPinger()
	pinger.Size = 64
	da, err := net.ResolveIPAddr("ip4:icmp", dest)
	if err != nil {
		return err
	}
	err = syscall.ETIMEDOUT
	pinger.AddIPAddr(da)
	pinger.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		fmt.Printf("%d bytes from %s in %s\n",
			pinger.Size, addr.String(), rtt.String())
		err = nil
	}
	pinger.OnIdle = func() {}
	fmt.Printf("PING %s (%s)\n", dest, da.String())
	if rerr := pinger.Run(); err == nil {
		err = rerr
	}
	return err
}

func (cmd) Man() lang.Alt  { return man }
func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)
