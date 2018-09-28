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

type Ifname [IFNAMSIZ]byte

func (ifname *Ifname) cache(args ...interface{}) {
	for _, v := range args {
		switch t := v.(type) {
		case *[IFNAMSIZ]uint8:
			copy(ifname[:], t[:])
		case *Ifname:
			copy(ifname[:], t[:])
		case string:
			copy(ifname[:], t)
			if len(t) < IFNAMSIZ {
				ifname[len(t)] = 0
			}
		}
	}
}

func (ifname *Ifname) String() string {
	for i, c := range ifname[:] {
		if c == 0 {
			return string(ifname[:i])
		}
	}
	return string(ifname[:])
}
