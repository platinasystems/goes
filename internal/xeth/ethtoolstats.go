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

import "fmt"

type EthtoolStat uint64

var EthtoolStats []string

func (stat EthtoolStat) String() string {
	var s string
	i := int(stat)
	if i < len(EthtoolStats) {
		s = EthtoolStats[i]
	} else {
		s = fmt.Sprint("@", i)
	}
	return s
}

var EthtoolStatMap map[string]EthtoolStat

func EthtoolStatOf(s string) (EthtoolStat, bool) {
	if EthtoolStatMap == nil {
		EthtoolStatMap = make(map[string]EthtoolStat)
		for i, s := range EthtoolStats {
			EthtoolStatMap[Hyphenate(s)] = EthtoolStat(i)
		}
	}
	ethtoolstat, found := EthtoolStatMap[Hyphenate(s)]
	return ethtoolstat, found
}
