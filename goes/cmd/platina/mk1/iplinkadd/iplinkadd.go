// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package iplinkadd

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/platinasystems/go/goes/cmd/eeprom"
	"github.com/platinasystems/go/goes/cmd/eeprom/platina_eeprom"
	"github.com/platinasystems/go/goes/cmd/ip"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/sockfile"
)

var IP = ip.Goes.Main

func init() {
	if false {
		IP = func(args ...string) error {
			ebuf := new(bytes.Buffer)
			cmd := exec.Command("ip", args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = ebuf
			cmd.Run()
			if ebuf.Len() > 0 {
				return errors.New(ebuf.String())
			}
			return nil
		}
	}
}

type Command struct{}

func (Command) String() string { return "platina" }

func (Command) Usage() string {
	return `
ip link add type platina eth-PORT-SUBPORT...`
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "add a platina virtual link",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
	PORT	{ 0..31 }
	SUBPORT	{ 0.. 3 }

SEE ALSO
	ip link add man type || ip link add type -man
	ip link man add || ip link add -man
	man ip || ip -man`,
	}
}

func (Command) Main(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing eth-PORT-SUBPORT...")
	}
	mkDev := mkDevVlan
	{
		fi, err := os.Stat("/sys/bus/pci/driver/ixgbevf")
		if err == nil && fi.IsDir() {
			mkDev = mkDevSriov
		}
	}
	ixgbes, err := getIxgbes()
	if err != nil {
		return err
	}
	eth0, err := net.InterfaceByName("eth0")
	if err != nil {
		return err
	}
	po, err := portOffset()
	if err != nil {
		return err
	}
	min := struct{ port, subport uint }{po + 0, po + 0}
	max := struct{ port, subport uint }{po + 31, po + 3}
	re := regexp.MustCompile("eth-(?P<port>[0-9]+)-(?P<subport>[0-9]+)")
	for _, name := range args {
		var port, subport uint
		match := re.FindStringSubmatch(name)
		if len(match) != 3 {
			return fmt.Errorf("%q: invalid name, expecting %q",
				name, "eth-PORT-SUBPORT")
		}
		fmt.Sscan(match[1], &port)
		if min.port > port || port > max.port {
			return fmt.Errorf("%q: port range { %d..%d }",
				name, min.port, max.port)
		}
		if po == 1 {
			port -= 1
		}
		fmt.Sscan(match[2], &subport)
		if min.subport > subport || subport > max.subport {
			return fmt.Errorf("%q: subport range { %d..%d }",
				name, min.subport, max.subport)
		}
		if po == 1 {
			subport -= 1
		}
		ixgbe := ixgbes[port&1]
		mac := macplus(eth0, 3+(port*4)+subport)
		vfi := ((port >> 1) * 4) + subport
		vf := fmt.Sprint(vfi)
		vlan := fmt.Sprint(1 + (4 * (port ^ 1)) + subport + 1)
		err := mkDev(name, ixgbe, vf, mac, vlan)
		if err != nil {
			return err
		}
	}
	return nil
}

func mkDevVlan(name, link, vf, mac, vlan string) error {
	return IP("link", "add",
		"link", link,
		"name", name,
		"address", mac,
		"type", "vlan",
		"id", vlan,
		"protocol", "802.1Q",
	)
}

func mkDevSriov(name, link, vf, mac, vlan string) error {
	err := IP("link", "set", link,
		"vf", vf,
		"address", mac,
		"vlan", vlan,
	)
	if err != nil {
		return err
	}
	return nil
}

func getIxgbes() ([]string, error) {
	const pat = "/sys/class/net/*/device/driver/module/drivers/pci:ixgbe"
	ixgbes, err := filepath.Glob(pat)
	if err != nil {
		return ixgbes, err
	}
	if len(ixgbes) != 2 {
		return ixgbes, fmt.Errorf("no IXGBE devices")
	}
	lns := make([]string, 2)
	for i := range ixgbes {
		// /sys/class/net/DEV/device is a symlink to the bus id so,
		// it's the best thing to sort on to have consistent interfaces
		dfn := strings.TrimSuffix(ixgbes[i],
			"/driver/module/drivers/pci:ixgbe")
		lns[i], _ = os.Readlink(dfn)
		ixgbes[i] = strings.TrimPrefix(strings.TrimSuffix(dfn,
			"/device"), "/sys/class/net/")
	}
	if filepath.Base(lns[0]) > filepath.Base(lns[1]) {
		ixgbes[0], ixgbes[1] = ixgbes[1], ixgbes[0]
	}
	return ixgbes, nil
}

func macplus(eth0 *net.Interface, u uint) string {
	mac := make(net.HardwareAddr, len(eth0.HardwareAddr))
	copy(mac[:], eth0.HardwareAddr[:])
	base := mac[5]
	mac[5] += byte(u)
	if mac[5] < base {
		base = mac[4]
		mac[4] += 1
		if mac[4] < base {
			mac[3] += 1
		}
	}
	return mac.String()
}

func portOffset() (po uint, err error) {
	const dv = "eeprom.DeviceVersion"
	var ver uint
	fi, err := os.Stat(sockfile.Path("redisd"))
	if err == nil && (fi.Mode()&os.ModeSocket) == os.ModeSocket {
		var s string
		s, err = redis.Hget(redis.DefaultHash, dv)
		if err == nil {
			_, err = fmt.Sscan(s, &ver)
		}
	} else {
		var p eeprom.Eeprom
		var buf []byte
		platina_eeprom.Config(
			platina_eeprom.BusIndex(0),
			platina_eeprom.BusAddress(0x51),
			platina_eeprom.BusDelay(10*time.Millisecond),
			platina_eeprom.MinMacs(132),
			platina_eeprom.OUI([3]byte{0x02, 0x46, 0x8a}),
		)
		buf, err = platina_eeprom.ReadBytes()
		if err == nil {
			if _, err = p.Write(buf); err == nil {
				v, found := p.Tlv[eeprom.DeviceVersionType]
				if !found {
					err = fmt.Errorf("not found")
				} else {
					ver = uint(v.(*eeprom.Hex8).Bytes()[0])
				}
			}
		}
	}
	if err != nil {
		err = fmt.Errorf("%s: %s", dv, err)
	} else if ver != 0 && ver != 0xff {
		po = 1
	}
	return
}
