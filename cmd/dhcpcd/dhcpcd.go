// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package dhcpcd is a simple dhcp client
package dhcpcd

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"math/bits"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"github.com/d2g/dhcp4"
	"github.com/d2g/dhcp4client"
	"github.com/jpillora/backoff"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/external/parms"
	"github.com/platinasystems/goes/lang"
)

type Command struct {
	g    *goes.Goes
	done chan struct{}
}

func (*Command) String() string { return "dhcpcd" }

func (*Command) Usage() string { return "dhcpcd" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "dhcp client daemon",
	}
}

func (c *Command) Close() error {
	close(c.done)
	return nil
}

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (*Command) Kind() cmd.Kind { return cmd.Daemon }

const IFNAMSIZ = 16

type ifreq struct {
	ifrName    [IFNAMSIZ]byte
	ifrNewname [IFNAMSIZ]byte
}

func (c *Command) parseACK(ack dhcp4.Packet) (myIP string, rtrIP string, dnsIP string, lt uint32, err error) {
	myIP = ack.YIAddr().String()
	opt := ack.ParseOptions()
	nm := opt[1]
	fmt.Printf("Option[1] netmask: %v\n", nm)
	if len(nm) == 4 {
		myIP = myIP + "/" + strconv.Itoa(bits.LeadingZeros32(^binary.BigEndian.Uint32(nm)))
	}
	fmt.Printf("Got address %s\n", myIP)

	rtr := opt[3]
	fmt.Printf("Option[3] router: %v\n", rtr)
	if len(rtr) == 4 {
		ip := net.IP(rtr)
		if !ip.Equal(net.IPv4(0, 0, 0, 0)) {
			rtrIP = ip.String()
		}
	}

	ltOpt := opt[51]
	fmt.Printf("Got lease %v\n", ltOpt)
	lt = uint32(86400)
	if len(ltOpt) == 4 {
		lt = binary.BigEndian.Uint32(ltOpt)
		fmt.Printf("Lease time %d\n", lt)
	}
	dns := opt[6]
	fmt.Printf("Got DNS %v\n", dns)
	for i := 0; i < len(dns) && len(dns[i:]) >= 4; i += 4 {
		dnsIP = dnsIP + "nameserver " + net.IP(dns[i:i+4]).String() + "\n"
	}
	fmt.Printf("DNS resolved to %v\n", dnsIP)

	return
}

func (c *Command) updateParm(dev string, myIP string, myLastIP string, rtrIP string, rtrLastIP string, dnsIP string, dnsLastIP string) (err error) {
	if myIP != myLastIP {
		if myIP != "" {
			err = c.g.Main("ip", "address", "add", myIP, "dev", dev)
			if err != nil {
				return err
			}
		}
		if myLastIP != "" {
			err = c.g.Main("ip", "address", "delete", myLastIP, "dev", dev)
			if err != nil {
				return err
			}
		}
	}
	if rtrIP != rtrLastIP {
		if rtrIP != "" {
			err := c.g.Main("ip", "route", "add", "0.0.0.0/0", "via", rtrIP)
			if err != nil {
				return err
			}
		}
		if rtrLastIP != "" {
			err := c.g.Main("ip", "route", "delete", "0.0.0.0/0", "via", rtrLastIP)
			if err != nil {
				return err
			}
		}
	}
	if dnsIP != dnsLastIP {
		if dnsIP != "" {
			err := ioutil.WriteFile("/etc/resolv.conf", []byte(dnsIP), 0644)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Command) Main(args ...string) error {
	parm, args := parms.New(args, "-i")
	i := "eth0"
	if parm.ByName["-i"] != "" {
		i = parm.ByName["-i"]
	}
	if len(i) > (IFNAMSIZ)-1 {
		return errors.New("Interface name too long")
	}

	var dev ifreq
	copy(dev.ifrName[:], i)

	s, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return err
	}
	defer syscall.Close(s)

	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(s),
		uintptr(syscall.SIOCGIFHWADDR), uintptr(unsafe.Pointer(&dev)))
	if e != 0 {
		return e
	}
	mac := net.HardwareAddr(dev.ifrNewname[2:8])

	fmt.Printf("Got %s\n", mac)

	c.done = make(chan struct{})

	err = c.g.Main("ip", "link", "change", i, "up")
	if err != nil {
		return err
	}
	defer func() {
		_ = c.g.Main("ip", "link", "change", i, "down")
	}()

	err = c.g.Main("ip", "route", "add", "255.255.255.255/32", "dev", i)
	if err != nil {
		return err
	}
	defer func() {
		_ = c.g.Main("ip", "route", "delete", "255.255.255.255/32", "dev", i)
	}()
	sock, err := dhcp4client.NewInetSock(dhcp4client.SetLocalAddr(net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 68}), dhcp4client.SetRemoteAddr(net.UDPAddr{IP: net.IPv4bcast, Port: 67}))
	if err != nil {
		return err
	}
	defer sock.Close()

	cl, err := dhcp4client.New(dhcp4client.HardwareAddr(mac), dhcp4client.Connection(sock))
	if err != nil {
		return err
	}
	defer cl.Close()

	b := &backoff.Backoff{
		Min:    1 * time.Second,
		Max:    60 * time.Second,
		Factor: 2,
		Jitter: false,
	}
	var myIP, rtrIP, dnsIP string
	var lt uint32
	var ack dhcp4.Packet
	var success bool

	for {
		success, ack, err = cl.Request()
		if err == nil {
			if success {
				myIP, rtrIP, dnsIP, lt, err = c.parseACK(ack)
				if err == nil {
					err := c.updateParm(i, myIP, "", rtrIP, "", dnsIP, "")
					if err == nil {
						break
					} else {
						fmt.Fprintf(os.Stderr, "Error in updateParm: %s\n", err)
					}
				} else {
					fmt.Fprintf(os.Stderr, "Error in parseACK: %s\n", err)
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error in Request: %s\n", err)
		}

		if !func() bool {
			t := time.NewTicker(b.Duration())
			defer t.Stop()

			select {
			case <-c.done:
				return false
			case <-t.C:
				return true
			}
		}() {
			return nil
		}
	}

	defer func() {
		c.updateParm(i, "", myIP, "", rtrIP, "", dnsIP)
	}()

	for {
		if !func() bool {
			t := time.NewTicker(time.Duration(lt) * time.Second)
			defer t.Stop()

			select {
			case <-c.done:
				return false
			case <-t.C:
				return true
			}
		}() {
			return nil
		}
		myLastIP := myIP
		rtrLastIP := rtrIP
		dnsLastIP := dnsIP

		for {
			success, ack, err = cl.Renew(ack)
			if err == nil {
				if success {
					myIP, rtrIP, dnsIP, lt, err = c.parseACK(ack)
					if err == nil {
						err = c.updateParm(i, myIP, myLastIP, rtrIP, rtrLastIP, dnsIP, dnsLastIP)
						if err == nil {
							break
						} else {
							fmt.Fprintf(os.Stderr, "Error in updateParm: %s\n", err)
						}
					} else {
						fmt.Fprintf(os.Stderr, "Error in parseACK: %s\n", err)
					}
				}
			} else {
				fmt.Fprintf(os.Stderr, "Error in Renew: %s\n", err)
			}
			if !func() bool {
				t := time.NewTicker(b.Duration())
				defer t.Stop()

				select {
				case <-c.done:
					return false
				case <-t.C:
					return true
				}
			}() {
				return nil
			}
		}
	}

	return nil
}
