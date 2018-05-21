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

type Err string

func (err Err) Error() string  { return string(err) }
func (err Err) String() string { return string(err) }

func (hdr *Hdr) IsHdr() bool {
	return hdr.Z64 == 0 && hdr.Z32 == 0 && hdr.Z16 == 0 && hdr.Z8 == 0
}

type BreakMsg struct {
	Hdr
}

type StatMsg struct {
	Hdr    Hdr
	Ifname Ifname
	Stat   Stat
}

func (msg *StatMsg) String() string {
	var stat fmt.Stringer
	op := Op(msg.Hdr.Op)
	switch op {
	case XETH_LINK_STAT_OP:
		stat = LinkStat(msg.Stat.Index)
	case XETH_ETHTOOL_STAT_OP:
		stat = EthtoolStat(msg.Stat.Index)
	default:
		stat = Err("unknown")
	}
	return fmt.Sprint(op, ": ", msg.Ifname, ": ", stat, ": ",
		msg.Stat.Count)
}

type EthtoolFlagsMsg struct {
	Hdr    Hdr
	Ifname Ifname
	Flags  EthtoolFlagBits
}

func (msg *EthtoolFlagsMsg) String() string {
	return fmt.Sprint(Op(msg.Hdr.Op), ": ", &msg.Ifname, ": ", msg.Flags)
}

type EthtoolSettingsMsg struct {
	Hdr      Hdr
	Ifname   Ifname
	Settings EthtoolSettings
}

func (msg *EthtoolSettingsMsg) String() string {
	return fmt.Sprint(Op(msg.Hdr.Op), ": ", &msg.Ifname, ":", &msg.Settings)
}

type CarrierMsg struct {
	Hdr    Hdr
	Ifname Ifname
	Flag   CarrierFlag
}

func (msg *CarrierMsg) String() string {
	return fmt.Sprint(Op(msg.Hdr.Op), ": ", &msg.Ifname, ":", msg.Flag)
}

type SpeedMsg struct {
	Hdr    Hdr
	Ifname Ifname
	Speed  Mbps
}

func (msg *SpeedMsg) String() string {
	return fmt.Sprint(Op(msg.Hdr.Op), ": ", &msg.Ifname, ":", msg.Speed)
}
