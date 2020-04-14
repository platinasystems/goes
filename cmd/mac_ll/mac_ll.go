// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mac_ll

import (
	"fmt"
	"net"

	"github.com/platinasystems/goes/external/redis"
	"github.com/platinasystems/goes/lang"
)

func New() Command { return Command{} }

type Command struct{}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: "convert mac addresss to IPv6 link-local",
	}
}
func (Command) String() string { return "mac-ll" }
func (Command) Usage() string  { return "mac-ll <mac>" }
func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION 
        Convert a 6 byte MAC address to an IPv6 link-local.

        Without any args, show the base MAC address plus link-local.

        <mac>    Convert the MAC address isIPv6 link-local.`,
	}
}

func (Command) Main(args ...string) error {
	var (
		mac, mac_hdr string
		hw           net.HardwareAddr
		err          error
	)

	if len(args) > 1 {
		return fmt.Errorf("%v: unexpected", args)
	}

	if len(args) == 0 {
		if mac, err = redis.Hget(redis.DefaultHash,
			"eeprom.BaseEthernetAddress"); err != nil {
			return err
		}
		mac_hdr = "Base MAC"
	} else {
		mac = args[0]
		mac_hdr = "MAC"
	}

	if hw, err = net.ParseMAC(mac); err != nil {
		return err
	}
	if len(hw) != 6 {
		return fmt.Errorf("%v: invalid 6 byte mac", hw)
	}

	buf := make([]byte, 16)
	buf[0] = 0xfe
	buf[1] = 0x80
	buf[8] = hw[0] ^ 0x02
	buf[9] = hw[1]
	buf[10] = hw[2]
	buf[11] = 0xff
	buf[12] = 0xfe
	buf[13] = hw[3]
	buf[14] = hw[4]
	buf[15] = hw[5]
	ll := net.IP(buf)

	fmt.Printf("%15s: %v\n", mac_hdr, mac)
	fmt.Printf("IPv6 link-local: %v\n", ll)

	return nil
}
