// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ping

import (
	"fmt"
	"net"
	"time"

	"github.com/tatsushid/go-fastping"
)

const Name = "ping"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " DESTINATION" }

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
	pinger.AddIPAddr(da)
	pinger.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		fmt.Printf("%d bytes from %s in %s\n",
			pinger.Size, addr.String(), rtt.String())
	}
	pinger.OnIdle = func() {}
	fmt.Printf("PING %s (%s)\n", dest, da.String())
	return pinger.Run()
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "send ICMP ECHO_REQUEST to network host",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	ping - send ICMP ECHO_REQUEST to network hosts

SYNOPSIS
	ping DESTINATION

DESCRIPTION
	Send ICMP ECHO_REQUEST to given host and print ECHO_REPLY.`,
	}
}
