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

type Supported EthtoolLinkModeBits
type Advertising EthtoolLinkModeBits
type Partner EthtoolLinkModeBits

type EthtoolSettings struct {
	Speed Mbps
	Autoneg
	Duplex
	DevPort
	Supported
	Advertising
	Partner
}

func (p *EthtoolSettings) cache(args ...interface{}) {
	for _, v := range args {
		switch t := v.(type) {
		case *MsgEthtoolSettings:
			p.Speed = Mbps(t.Speed)
			p.Duplex = Duplex(t.Duplex)
			p.DevPort = DevPort(t.Port)
			p.Supported = *(*Supported)(&t.Link_modes_supported)
			p.Advertising =
				*(*Advertising)(&t.Link_modes_advertising)
			p.Partner = *(*Partner)(&t.Link_modes_lp_advertising)
		case Mbps:
			p.Speed = t
		case Duplex:
			p.Duplex = t
		case DevPort:
			p.DevPort = t
		case Autoneg:
			p.Autoneg = t
		case *Supported:
			p.Supported = *t
		case *Advertising:
			p.Advertising = *t
		case *Partner:
			p.Partner = *t
		}
	}
}
