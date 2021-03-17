// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package install

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

var networkSetupScript = `auto {{ .MgmtEth }}

iface {{ .MgmtEth }} inet static
	address {{ .MgmtIP }}
	gateway {{ .MgmtGW }}
`

func (c *Command) networkSetup() (err error) {
	d, err := ioutil.ReadFile("/etc/resolv.conf")
	if err == nil {
		err = ioutil.WriteFile(filepath.Join(c.Target, "etc/resolv.conf"),
			d, 0644)
		if err != nil {
			return fmt.Errorf("networkSetup: Error writing /etc/resolv.conf: %w",
				err)
		}
	}

	err = c.writeTemplateToFile(c.MgmtEth, networkSetupScript)
	if err != nil {
		err = fmt.Errorf("networkSetup: Error writing %s: %w",
			c.MgmtEth, err)
	}
	return err
}

func ipFromInterface(ifaceName string, v6 bool) (ipnet *net.IPNet, err error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil, fmt.Errorf("net.InterfaceByName(%s) failed: %w",
			ifaceName, err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("iface.Addrs(%s) failed: %w",
			ifaceName, err)
	}
	for _, addr := range addrs {
		var ok bool
		if ipnet, ok = addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			ipAddr := ipnet.IP
			if ipv4 := ipAddr.To4(); ipv4 != nil {
				if !v6 {
					return
				}
			} else {
				if v6 {
					return
				}
			}
		}
	}
	return nil, nil
}

func defaultGateway() (gwIP net.IP, iface string, err error) {
	f, err := os.Open("/proc/net/route")
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)

	_ = scanner.Scan() // Skip header line
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) != 11 {
			return nil, "", fmt.Errorf("Unexpected entry in /proc/net/route: %s", line)
		}
		dest, err := strconv.ParseUint(fields[1], 16, 64)
		if err != nil {
			return nil, "", fmt.Errorf("Error parsing /proc/net/route dest %s: %w",
				fields[1], err)
		}

		gw, err := strconv.ParseUint(fields[2], 16, 64)
		if err != nil {
			return nil, "", fmt.Errorf("Error parsing /proc/net/route gw %s: %w",
				fields[2], err)
		}

		mask, err := strconv.ParseUint(fields[7], 16, 64)
		if err != nil {
			return nil, "", fmt.Errorf("Error parsing /proc/net/route mask %s: %w",
				fields[7], err)
		}

		flags, err := strconv.ParseUint(fields[3], 16, 64)
		if err != nil {
			return nil, "", fmt.Errorf("Error parsing /proc/net/route flags %s: %w",
				fields[3], err)
		}

		if dest != 0 || mask != 0 || !(flags&syscall.RTF_GATEWAY != 0) {
			continue
		}

		g4 := byte(gw >> 24)
		g3 := byte(gw >> 16)
		g2 := byte(gw >> 8)
		g1 := byte(gw)

		gwIP = net.IPv4(g1, g2, g3, g4)
		return gwIP, fields[0], nil
	}
	return nil, "", fmt.Errorf("Default IPv4 route not found")
}

func defaultGatewayV6() (gwIP net.IP, iface string, err error) {
	f, err := os.Open("/proc/net/ipv6_route")
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) != 10 {
			return nil, "", fmt.Errorf("Unexpected entry in /proc/net/ipv6_route: %s",
				line)
		}
		if len(fields[0]) != 32 {
			return nil, "", fmt.Errorf("Unexpected format parsing /proc/net/ipv6_route dest: %s",
				fields[0])
		}
		dest1, err := strconv.ParseUint(fields[0][0:15], 16, 64)
		if err != nil {
			return nil, "", fmt.Errorf("Error parsing /proc/net/ipv6_route dest %s: %w",
				fields[0], err)
		}

		dest2, err := strconv.ParseUint(fields[0][16:31], 16, 64)
		if err != nil {
			return nil, "", fmt.Errorf("Error parsing /proc/net/ipv6_route dest %s: %w",
				fields[0], err)
		}

		pfxlen, err := strconv.ParseUint(fields[1], 16, 64)
		if err != nil {
			return nil, "", fmt.Errorf("Error parsing /proc/net/ipv6_prefix len %s: %w",
				fields[1], err)
		}

		gw, err := hex.DecodeString(fields[4])
		if err != nil {
			return nil, "", fmt.Errorf("Error parsing gw %s: %w",
				fields[4], err)
		}
		if len(gw) != 16 {
			return nil, "", fmt.Errorf("Unexpected format parsing /proc/net/ipv6_route next hop: %s",
				fields[4])
		}

		flags, err := strconv.ParseUint(fields[8], 16, 64)
		if err != nil {
			return nil, "", fmt.Errorf("Error parsing /proc/net/ipv6_route flags %s: %w",
				fields[8], err)
		}

		if dest1 != 0 || dest2 != 0 || pfxlen != 0 ||
			!(flags&syscall.RTF_GATEWAY != 0) {
			continue
		}

		return gw, fields[9], nil
	}
	return nil, "", fmt.Errorf("Default IPv6 route not found")
}

func (c *Command) updateDNS() {
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		fmt.Printf("Unable to parse /etc/resolv.conf: %s\n", err)
		return
	}
	defer file.Close()

	ns := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		l := scanner.Text()
		f := strings.Fields(l)
		if len(f) != 2 || f[0] != "nameserver" {
			continue
		}
		if ns != "" {
			ns = ns + " "
		}
		ns = ns + f[1]
	}
	c.DefaultDNS = ns
}
