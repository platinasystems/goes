// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ping

import (
	"fmt"
	"net"
	"syscall"
	"time"

	"github.com/platinasystems/goes/lang"
	"github.com/tatsushid/go-fastping"
)

type Command struct{}

func (Command) String() string { return "ping" }

func (Command) Usage() string {
	return "ping DESTINATION"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "send ICMP ECHO_REQUEST to network host",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Send ICMP ECHO_REQUEST to given host and print ECHO_REPLY.`,
	}
}

func (Command) Main(args ...string) error {
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
