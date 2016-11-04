// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"

	"fmt"
)

type l3_main struct {
	*fe1a
	my_station_tcam my_station_tcam_main
}

func (t *fe1a) register_hooks(v *vnet.Vnet) {
	t.l3_interface_init()
	t.flex_counter_init()
	t.adjacency_main_init()
	t.ip4_fib_init()
	v.RegisterSwIfAdminUpDownHook(t.l3_main.swIfAdminUpDown)
	v.RegisterSwIfAddDelHook(t.swIfAddDel)
	v.RegisterSwIfCounterSyncHook(t.swIfCounterSync)
}

func port_for_si(v *vnet.Vnet, si vnet.Si) (p *Port) {
	h, ok := v.HwIferForSi(si)
	if ok {
		p, ok = h.(*Port)
	}
	return
}

func (lm *l3_main) swIfAdminUpDown(v *vnet.Vnet, si vnet.Si, isUp bool) (err error) {
	p := port_for_si(v, si)
	if p == nil {
		return
	}

	e := my_station_tcam_entry{}
	e.key.LogicalPort.Set(uint(p.physical_port_number.toGpp()))
	e.mask.LogicalPort = m.LogicalPortMaskAll
	e.key.EthernetAddress = m.EthernetAddress(p.Address)
	e.mask.EthernetAddress = m.EthernetAddressMaskAll
	e.data = my_station_tcam_data{
		ip4_unicast_enable:   true,
		ip6_unicast_enable:   true,
		ip4_multicast_enable: true,
		ip6_multicast_enable: true,
		mpls_enable:          true,
		arp_rarp_enable:      true,
	}
	isDel := !isUp
	lm.my_station_tcam.addDel(lm.fe1a, &e, isDel)
	return
}

var rx_pipe_debug_events_0 = [...]string{
	31: "parity error packet",
	30: "LAG failover lopoback packet dropped; backup port down",
	29: "LAG failover loopback packet from loopback/hi-gig",
	28: "multicast index error packet",
	27: "hi-gig header error packets not copied to CPU",
	26: "packets dropped due to hi-gig failover port down",
	25: "tunnel error packets",
	23: "L3 MTU exceeded; packet sent to CPU",
	22: "DOS fragment error packets",
	21: "DOS ICMP error packets",
	20: "DOS L4 header error packets",
	19: "DOS L3 header error packets",
	18: "DOS L2 header error packets",
	16: "VFP drop",
	15: "port bitmap zero drop",
	14: "multicast L2+L3 packets dropped",
	13: "IFP packets dropped",
	12: "policy discard - dst_discard, src_discard, rate_control, etc.",
	11: "invalid VLAN packets dropped",
	10: "packets dropped in discard stage",
	9:  "packets dropped with hi-gig header type 1",
	8:  "packets dropped due to CFI or L3 disable",
	7:  "packets dropped due to L2 destination discard",
	6:  "packets dropped due to L2 source static movement",
	5:  "packets dropped due to L2 source discard",
	4:  "packets dropped due to source route",
	3:  "packets dropped due to CML",
	2:  "protocol packet drop",
	1:  "BPDU packet drop",
	0:  "packets dropped due to VLAN translation miss",
}

var rx_pipe_debug_events_1 = [...]string{
	19: "time sync packets dropped",
	17: "L2/L3 DST_DISCARD is set",
	16: "URPF check failed",
	15: "VLAN drop",
	5:  "MPLS drop: label miss, invalid action, invalid payload, TTL failure",
	4:  "unallowed class-based station move",
	3:  "L2 non-uc drop or VLAN cross-connect miss or L2 miss drop or VLAN cross-connect with wrong dst_type",
	2:  "source ethernet address is zero",
	1:  "LAG failover packet detected on loopback port - treat as ingress mirror copy (so switched copy is dropped)",
	0:  "hi-gig header errors - whether or not copied to cpu",
}

var rx_pipe_debug_events_2 = [...]string{
	2: "unknown BFD control packet",
	1: "BFD version number of ACH header non-zero or ACH channel type is unknown",
	0: "BFD version number not supported, or session lookup failure",
}

func (ss *switchSelect) show_rx_tx_pipe_debug_events(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	ss.SelectFromInput(in)
	for _, sw := range ss.Switches {
		t := sw.(*fe1a)
		q := t.getDmaReq()
		nEvents := 0
		for pipe := uint(0); pipe < 4; pipe++ {
			var (
				v [3]uint32
				r [1]uint32
			)
			for i := range v {
				t.rx_pipe_regs.event_debug[i].geta(q, sbus.Unique(pipe), &v[i])
				t.rx_pipe_regs.event_debug[i].seta(q, sbus.Unique(pipe), 0)
			}
			for i := range r {
				t.tx_pipe_regs.event_debug.geta(q, sbus.Unique(pipe), &r[i])
				t.tx_pipe_regs.event_debug.seta(q, sbus.Unique(pipe), 0)
			}
			q.Do()

			if v[0] != 0 {
				for i := range rx_pipe_debug_events_0 {
					if s := rx_pipe_debug_events_0[i]; len(s) > 0 && v[0]&(1<<uint(i)) != 0 {
						fmt.Fprintf(w, "rx pipe %d: %s\n", pipe, s)
						nEvents++
					}
				}
			}
			if v[1] != 0 {
				for i := range rx_pipe_debug_events_1 {
					if s := rx_pipe_debug_events_1[i]; len(s) > 0 && v[0]&(1<<uint(i)) != 0 {
						fmt.Fprintf(w, "rx pipe %d: %s\n", pipe, s)
						nEvents++
					}
				}
			}
			if v[2] != 0 {
				for i := range rx_pipe_debug_events_2 {
					if s := rx_pipe_debug_events_2[i]; len(s) > 0 && v[0]&(1<<uint(i)) != 0 {
						fmt.Fprintf(w, "rx pipe %d: %s\n", pipe, s)
						nEvents++
					}
				}
			}

			if r[0] != 0 {
				fmt.Fprintf(w, "tx pipe %d: 0x%x\n", pipe, v, r[0])
				nEvents++
			}
		}
		if nEvents == 0 {
			fmt.Fprintln(w, "no debug events")
		}
	}
	return
}
