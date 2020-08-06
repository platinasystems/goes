// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
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
	g     *goes.Goes
	myIP  string
	rtrIP string
	dnsIP string
	lt    uint32
	ack   dhcp4.Packet
	cl    *dhcp4client.Client
	i     string
}

func (*Command) String() string { return "dhcpcd" }

func (*Command) Usage() string { return "dhcpcd" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "dhcp client daemon",
	}
}

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (*Command) Kind() cmd.Kind { return cmd.Daemon }

const IFNAMSIZ = 16

type ifreq struct {
	ifrName    [IFNAMSIZ]byte
	ifrNewname [IFNAMSIZ]byte
}

func (c *Command) parseACK(ack dhcp4.Packet) (err error) {
	c.myIP = ack.YIAddr().String()
	opt := ack.ParseOptions()
	nm := opt[1]
	fmt.Printf("Option[1] netmask: %v\n", nm)
	if len(nm) == 4 {
		c.myIP = c.myIP + "/" + strconv.Itoa(bits.LeadingZeros32(^binary.BigEndian.Uint32(nm)))
	}
	fmt.Printf("Got address %s\n", c.myIP)

	rtr := opt[3]
	c.rtrIP = ""
	fmt.Printf("Option[3] router: %v\n", rtr)
	if len(rtr) == 4 {
		ip := net.IP(rtr)
		if !ip.Equal(net.IPv4(0, 0, 0, 0)) {
			c.rtrIP = ip.String()
		}
	}

	ltOpt := opt[51]
	fmt.Printf("Got lease %v\n", ltOpt)
	c.lt = uint32(86400)
	if len(ltOpt) == 4 {
		c.lt = binary.BigEndian.Uint32(ltOpt)
		fmt.Printf("Lease time %d\n", c.lt)
	}
	dns := opt[6]
	fmt.Printf("Got DNS %v\n", dns)
	c.dnsIP = ""
	for i := 0; i < len(dns) && len(dns[i:]) >= 4; i += 4 {
		c.dnsIP = c.dnsIP + "nameserver " + net.IP(dns[i:i+4]).String() + "\n"
	}
	fmt.Printf("DNS resolved to %v\n", c.dnsIP)

	return
}

func (c *Command) updateParm(myIP string, myLastIP string, rtrIP string,
	rtrLastIP string, dnsIP string, dnsLastIP string) (err error) {
	if myIP != myLastIP {
		if myIP != "" {
			err = c.g.Main("ip", "address", "add", myIP, "dev", c.i)
			if err != nil {
				return err
			}
		}
		if myLastIP != "" {
			err = c.g.Main("ip", "address", "delete", myLastIP, "dev", c.i)
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
	c.i = "eth0"
	if parm.ByName["-i"] != "" {
		c.i = parm.ByName["-i"]
	}
	if len(c.i) > (IFNAMSIZ)-1 {
		return errors.New("Interface name too long")
	}

	var dev ifreq
	copy(dev.ifrName[:], c.i)

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

	err = c.g.Main("ip", "link", "change", c.i, "up")
	if err != nil {
		return err
	}
	defer func() {
		_ = c.g.Main("ip", "link", "change", c.i, "down")
	}()

	err = c.g.Main("ip", "route", "add", "255.255.255.255/32", "dev", c.i)
	if err != nil {
		return err
	}
	defer func() {
		_ = c.g.Main("ip", "route", "delete", "255.255.255.255/32", "dev", c.i)
	}()
	sock, err := dhcp4client.NewInetSock(dhcp4client.SetLocalAddr(net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 68}), dhcp4client.SetRemoteAddr(net.UDPAddr{IP: net.IPv4bcast, Port: 67}))
	if err != nil {
		return err
	}
	defer sock.Close()

	c.cl, err = dhcp4client.New(dhcp4client.HardwareAddr(mac), dhcp4client.Connection(sock))
	if err != nil {
		return err
	}
	defer c.cl.Close()

	defer func() {
		if c.ack != nil && c.myIP != "" {
			err := c.cl.Release(c.ack)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error in release: %s\n", err)
			}
		}
		c.updateParm("", c.myIP, "", c.rtrIP, "", c.dnsIP)
		c.myIP = ""
		c.rtrIP = ""
		c.dnsIP = ""
	}()

	b := &backoff.Backoff{
		Min:    1 * time.Second,
		Max:    60 * time.Second,
		Factor: 2,
		Jitter: false,
	}

	for {
		myLastIP := c.myIP
		rtrLastIP := c.rtrIP
		dnsLastIP := c.dnsIP
		success := false
		done := make(chan struct{}, 1)
		goes.WG.Add(1)
		go func() {
			defer goes.WG.Done()
			success, c.ack, err = c.cl.Request()
			close(done)
		}()
		select {
		case <-goes.Stop:
			return nil
		case <-done:
		}
		if err == nil {
			if success {
				err := c.parseACK(c.ack)
				if err == nil {
					if c.myIP != "" {
						err := c.updateParm(c.myIP, myLastIP, c.rtrIP, rtrLastIP,
							c.dnsIP, dnsLastIP)
						if err == nil {
							exit, err := c.renew()
							if exit {
								return nil
							}
							fmt.Fprintf(os.Stderr, "Error in renew: %s\n", err)
						} else {
							fmt.Fprintf(os.Stderr, "Error in updateParm: %s\n", err)
						}
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
			case <-goes.Stop:
				return false
			case <-t.C:
				return true
			}
		}() {
			return nil
		}
	}

	return nil
}

func (c *Command) renew() (done bool, err error) {
	timeout := time.Now().Add(time.Duration(c.lt) * time.Second)
	sleepTime := c.lt / 2
	for time.Now().Before(timeout) {
		if !func() bool {
			t := time.NewTicker(time.Duration(sleepTime) * time.Second)
			defer t.Stop()

			select {
			case <-goes.Stop:
				return false
			case <-t.C:
				return true
			}
		}() {
			return true, nil
		}
		sleepTime = sleepTime / 2
		if sleepTime < 1 {
			sleepTime = 1
		}
		myLastIP := c.myIP
		rtrLastIP := c.rtrIP
		dnsLastIP := c.dnsIP

		success := false
		done := make(chan struct{}, 1)
		goes.WG.Add(1)
		go func() {
			defer goes.WG.Done()
			success, c.ack, err = c.cl.Request()
			close(done)
		}()
		select {
		case <-goes.Stop:
			return true, nil
		case <-done:
		}
		if err == nil {
			if success {
				err := c.parseACK(c.ack)
				if err == nil {
					if c.myIP == "" {
						return false,
							fmt.Errorf("Renew did not contain IP address")
					}
					err = c.updateParm(c.myIP, myLastIP, c.rtrIP, rtrLastIP, c.dnsIP, dnsLastIP)
					if err == nil {
						timeout = time.Now().Add(time.Duration(c.lt) * time.Second)
						sleepTime = c.lt / 2
						continue
					} else {
						return false,
							fmt.Errorf("Error in updateParm: %w", err)
					}
				} else {
					return false,
						fmt.Errorf("Error in parseACK: %w", err)
				}
			}
		} else {
			return false, fmt.Errorf("Error in Renew: %w", err)
		}
	}
	return false, fmt.Errorf("Lease expired without renew")
}
