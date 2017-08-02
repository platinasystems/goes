// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/vnet"
)

type vnet_tun_main struct {
	m *Main
	vnet.SwInterfaceType
}

func (m *vnet_tun_main) SwInterfaceName(v *vnet.Vnet, s *vnet.SwIf) string {
	ns := m.m.net_namespace_main.namespace_pool.entries[s.GetId()]
	return "vnet-" + ns.name
}

func (m *vnet_tun_main) init(um *Main) {
	m.m = um
	um.v.RegisterSwInterfaceType(m)
}

func (m *vnet_tun_main) si_is_vnet_tun(si vnet.Si) bool {
	return si.Kind(m.m.v) == m.SwIfKind
}
func IsVnetTun(v *vnet.Vnet, si vnet.Si) bool {
	return GetMain(v).si_is_vnet_tun(si)
}

func (m *vnet_tun_main) create_tun(ns *net_namespace) (intf *tuntap_interface) {
	si := m.m.v.NewSwIf(m.SwIfKind, vnet.IfId(ns.index))
	intf = m.m.vnet_tuntap_interface_by_si[si]
	intf.namespace = ns
	return
}
