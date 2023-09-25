// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

const (
	IFLA_VF_STATS_RX_PACKETS = iota
	IFLA_VF_STATS_TX_PACKETS
	IFLA_VF_STATS_RX_BYTES
	IFLA_VF_STATS_TX_BYTES
	IFLA_VF_STATS_BROADCAST
	IFLA_VF_STATS_MULTICAST
	IFLA_VF_STATS_PAD
	N_IFLA_VF_STATS
)

const IFLA_VF_STATS_MAX = N_IFLA_VF_STATS - 1

type IflaVfStats [N_IFLA_VF_STATS][]byte
