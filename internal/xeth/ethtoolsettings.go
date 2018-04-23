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
	"bytes"
	"fmt"
)

type EthtoolSettings struct {
	cmd                 uint32
	Speed               uint32
	Duplex              Duplex
	Port                Port
	PhyAddress          uint8
	Autoneg             Autoneg
	MdioSupport         uint8
	EthTpMdix           uint8
	EthTpMdixCtrl       uint8
	LinkModeMasksNwords uint8
	reserved            [8]uint32
	Supported           EthtoolLinkModeBits
	Advertising         EthtoolLinkModeBits
	Partner             EthtoolLinkModeBits
}

func (settings *EthtoolSettings) String() string {
	buf := new(bytes.Buffer)
	if settings.Speed != 0 {
		fmt.Fprint(buf, "\n\tspeed: ", settings.Speed, "Mb/s")
	} else {
		fmt.Fprint(buf, "\n\tspeed: unspecified")
	}
	fmt.Fprint(buf, "\n\tduplex: ", settings.Duplex)
	fmt.Fprint(buf, "\n\tport: ", settings.Port)
	fmt.Fprint(buf, "\n\tautoneg: ", settings.Autoneg)
	fmt.Fprint(buf, "\n\tsupported link modes:", &settings.Supported)
	fmt.Fprint(buf, "\n\tadvertising link modes:", &settings.Advertising)
	fmt.Fprint(buf, "\n\tpartner link modes:", &settings.Partner)
	return buf.String()
}
