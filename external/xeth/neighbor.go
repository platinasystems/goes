// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xeth

import (
	"bytes"
	"net"
	"sync"
	"syscall"

	"github.com/platinasystems/goes/external/xeth/internal"
)

type Neighbor struct {
	NetNs
	Xid
	net.IP
	net.HardwareAddr
	Ref
}

var poolNeighbor = sync.Pool{
	New: func() interface{} {
		return &Neighbor{
			IP: make([]byte, net.IPv6len, net.IPv6len),

			HardwareAddr: make([]byte, internal.SizeofEthAddr),
		}
	},
}

func newNeighbor() *Neighbor {
	neigh := poolNeighbor.Get().(*Neighbor)
	neigh.Hold()
	return neigh
}

func (neigh *Neighbor) Pool() {
	if neigh.Release() == 0 {
		neigh.IP = neigh.IP[:net.IPv6len]
		poolNeighbor.Put(neigh)
	}
}

// to sort a list of neighbors,
//	sort.Slice(neighbors, func(i, j int) bool {
//		return neighbors[i].Less(neighbors[j])
//	})
func (neighI *Neighbor) Less(neighJ *Neighbor) bool {
	return bytes.Compare(neighI.IP, neighJ.IP) < 0
}

func neighbor(msg *internal.MsgNeighUpdate) *Neighbor {
	neigh := newNeighbor()
	netns := NetNs(msg.Net)
	neigh.NetNs = netns
	neigh.Xid = netns.Xid(msg.Ifindex)
	if msg.Family == syscall.AF_INET {
		copy(neigh.IP, msg.Dst[:net.IPv4len])
		neigh.IP = neigh.IP[:net.IPv4len]
	} else {
		copy(neigh.IP, msg.Dst[:int(msg.Len)])
		neigh.IP = neigh.IP[:net.IPv6len]
	}
	copy(neigh.HardwareAddr, msg.Lladdr[:])
	netns.neighbor(neigh)
	return neigh
}
