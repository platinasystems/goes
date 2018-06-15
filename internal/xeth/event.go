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

const (
	NETDEV_UP = 1 + iota
	NETDEV_DOWN
	NETDEV_REBOOT
	NETDEV_CHANGE
	NETDEV_REGISTER
	NETDEV_UNREGISTER
	NETDEV_CHANGEMTU
	NETDEV_CHANGEADDR
	NETDEV_GOING_DOWN
	NETDEV_CHANGENAME
	NETDEV_FEAT_CHANGE
	NETDEV_BONDING_FAILOVER
	NETDEV_PRE_UP
	NETDEV_PRE_TYPE_CHANGE
	NETDEV_POST_TYPE_CHANGE
	NETDEV_POST_INIT
	NETDEV_UNREGISTER_FINAL
	NETDEV_RELEASE
	NETDEV_NOTIFY_PEERS
	NETDEV_JOIN
	NETDEV_CHANGEUPPER
	NETDEV_RESEND_IGMP
	NETDEV_PRECHANGEMTU
	NETDEV_CHANGEINFODATA
	NETDEV_BONDING_INFO
	NETDEV_PRECHANGEUPPER
	NETDEV_CHANGELOWERSTATE
	NETDEV_UDP_TUNNEL_PUSH_INFO
	NETDEV_CHANGE_TX_QUEUE_LEN
)

type Event int

func (event Event) String() string {
	var events = []string{
		"reserved",
		"up",
		"down",
		"reboot",
		"change",
		"register",
		"unregister",
		"changemtu",
		"changeaddr",
		"going_down",
		"changename",
		"feat-change",
		"bonding-failover",
		"pre-up",
		"pre-type-change",
		"post-type-change",
		"post-init",
		"unregister-final",
		"release",
		"notify-peers",
		"join",
		"changeupper",
		"resend-igmp",
		"prechangemtu",
		"changeinfodata",
		"bonding-info",
		"prechangeupper",
		"changelowerstate",
		"udp-tunnel-push-info",
		"change-tx-queue-len",
	}
	i := int(event)
	if i < len(events) {
		return events[i]
	}
	return fmt.Sprint("@", i)
}
