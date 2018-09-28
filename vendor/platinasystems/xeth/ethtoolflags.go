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

type EthtoolPrivFlags uint32

var EthtoolPrivFlagNames []string

func (bits *EthtoolPrivFlags) cache(args ...interface{}) {
	for _, v := range args {
		switch t := v.(type) {
		case uint32:
			*bits = EthtoolPrivFlags(t)
		case EthtoolPrivFlags:
			*bits = t
		case *MsgEthtoolFlags:
			*bits = EthtoolPrivFlags(t.Flags)
		}
	}
}

func (bits EthtoolPrivFlags) String() string {
	var sep string
	if bits == 0 {
		return "none"
	}
	if len(EthtoolPrivFlagNames) == 0 {
		return fmt.Sprintf("b%b", bits)
	}
	buf := new(bytes.Buffer)
	for i, s := range EthtoolPrivFlagNames {
		bit := EthtoolPrivFlags(1) << uint(i)
		if bits&bit == bit {
			fmt.Fprint(buf, sep, s)
			sep = "|"
		}
	}
	return buf.String()
}

func (bits EthtoolPrivFlags) Test(bit uint) bool {
	mask := EthtoolPrivFlags(1 << bit)
	return (bits & mask) == mask
}
