/* Copyright(c) 2018 Platina Systems, Inc.
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms and conditions of the GNU General Public License,
 * version 2, as published by the Free Software Foundation.
 *
 * This program is distributed in the hope it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
 * more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * this program; if not, write to the Free Software Foundation, Inc.,
 * 51 Franklin St - Fifth Floor, Boston, MA 02110-1301 USA.
 *
 * The full GNU General Public License is included in this distribution in
 * the file called "COPYING".
 *
 * Contact Information:
 * sw@platina.com
 * Platina Systems, 3180 Del La Cruz Blvd, Santa Clara, CA 95054
 */
package xeth

import (
	"fmt"
	"strings"
)

type Ifname [IFNAMSIZ]byte

func (ifname *Ifname) String() string {
	for i, c := range ifname[:] {
		if c == 0 {
			return string(ifname[:i])
		}
	}
	return string(ifname[:])
}

// all ports must have same PortPrefix so qsfp can hset
// FIXME: remove SetPortPrefix() when mk1 only supports xeth
var prefixIsXeth bool
var PortPrefix string

func SetPortPrefix(ifname string) {
	var isXeth bool

	if strings.HasPrefix(ifname, "xeth") {
		isXeth = true
	} else if strings.HasPrefix(ifname, "eth-") {
		isXeth = false
	} else {
		panic(fmt.Sprintf("Invalid PortPrefix, ifname %v\n", ifname))
	}

	if len(PortPrefix) == 0 {
		if isXeth {
			prefixIsXeth = true
			PortPrefix = "xeth"
		} else {
			prefixIsXeth = false
			PortPrefix = "eth-"
		}
	} else if prefixIsXeth != isXeth {
		panic(fmt.Sprintf("Cannot change PortPrefix, ifname %v\n", ifname))
	}
}

func PortName(frontPanelIndex int, subPortIndex int) (ifname string) {

	if prefixIsXeth && subPortIndex == 0 {
		ifname = fmt.Sprintf("%v%v", PortPrefix, frontPanelIndex)
	} else {
		ifname = fmt.Sprintf("%v%v-%v", PortPrefix, frontPanelIndex, subPortIndex)
	}
	return
}
