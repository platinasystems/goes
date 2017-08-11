// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import "unsafe"

const (
	Rx_packets = iota
	Tx_packets /* total packets transmitted	*/
	Rx_bytes   /* total bytes received 	*/
	Tx_bytes   /* total bytes transmitted	*/
	Rx_errors  /* bad packets received		*/
	Tx_errors  /* packet transmit problems	*/
	Rx_dropped /* no space in linux buffers	*/
	Tx_dropped /* no space available in linux	*/
	Multicast  /* multicast packets received	*/
	Collisions
	Rx_length_errors
	Rx_over_errors   /* receiver ring buff overflow	*/
	Rx_crc_errors    /* recved pkt with crc error	*/
	Rx_frame_errors  /* recv'd frame alignment error */
	Rx_fifo_errors   /* recv'r fifo overrun		*/
	Rx_missed_errors /* receiver missed packet	*/
	Tx_aborted_errors
	Tx_carrier_errors
	Tx_fifo_errors
	Tx_heartbeat_errors
	Tx_window_errors
	Rx_compressed
	Tx_compressed
	N_link_stat
)

const SizeofIfStats = N_link_stat * 4
const SizeofIfStats64 = N_link_stat * 8

func IfStatsAttr(b []byte) *IfStats {
	return (*IfStats)(unsafe.Pointer(&b[0]))
}

func IfStats64Attr(b []byte) *IfStats64 {
	return (*IfStats64)(unsafe.Pointer(&b[0]))
}

type IfStats [N_link_stat]uint32
type IfStats64 [N_link_stat]uint64
