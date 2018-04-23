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

type LinkStat uint64

var LinkStats = []string{
	"rx-packets",
	"tx-packets",
	"rx-bytes",
	"tx-bytes",
	"rx-errors",
	"tx-errors",
	"rx-dropped",
	"tx-dropped",
	"multicast",
	"collisions",
	"rx-length-errors",
	"rx-over-errors",
	"rx-crc-errors",
	"rx-frame-errors",
	"rx-fifo-errors",
	"rx-missed-errors",
	"tx-aborted-errors",
	"tx-carrier-errors",
	"tx-fifo-errors",
	"tx-heartbeat-errors",
	"tx-window-errors",
	"rx-compressed",
	"tx-compressed",
	"rx-nohandler",
}

func (stat LinkStat) String() string {
	var s string
	i := int(stat)
	if i < len(LinkStats) {
		s = LinkStats[i]
	} else {
		s = fmt.Sprint("@", i)
	}
	return s
}

var LinkStatMap map[string]LinkStat

func LinkStatOf(s string) (LinkStat, bool) {
	if LinkStatMap == nil {
		LinkStatMap = make(map[string]LinkStat)
		for i, s := range LinkStats {
			LinkStatMap[Hyphenate(s)] = LinkStat(i)
		}
	}
	linkstat, found := LinkStatMap[Hyphenate(s)]
	return linkstat, found
}
