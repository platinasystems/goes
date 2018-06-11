// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux

package unix

import (
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/iomux"
	"github.com/platinasystems/go/internal/netlink"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ip4"

	"sync"
)

type tuntap_interface struct {
	m  *Main
	mu sync.Mutex
	// Namespace this interface is currently in.
	namespace *net_namespace
	// Raw socket bound to this interface used for provisioning.
	provision_fd   int
	dev_net_tun_fd int
	iomux.File     // /dev/net/tun fd for this interface.
	hi             vnet.Hi
	si             vnet.Si
	name           ifreq_name
	elog_name      elog.StringRef
	ifindex        uint32 // linux interface index

	// Tun (ip4/ip6 header) versus tap (has ethernet header).
	isTun bool

	// Tuntap interface has been created (via TUNSETIFF ioctl).
	created bool
	// True when vnet/kernel interface flag sync has started.
	flag_sync_in_progress bool
	// True when vnet/kernel interface flags have been successfully synchronized.
	flag_sync_done bool
	flags          iff_flag
	operState      netlink.IfOperState

	mtuBytes   uint
	mtuBuffers uint

	tuntap_interface_tx_node
	tuntap_interface_rx_node

	interface_routes ip4.MapFib
}

//go:generate gentemplate -d Package=unix -id ifVec -d VecType=interfaceVec -d Type=*tuntap_interface github.com/platinasystems/go/elib/vec.tmpl

func (i *tuntap_interface) Name() string   { return i.name.String() }
func (i *tuntap_interface) String() string { return i.Name() }

type tuntap_main struct {
	// Selects whether we create tun or tap interfaces.
	mtuBytes   uint
	bufferPool *vnet.BufferPool
}
