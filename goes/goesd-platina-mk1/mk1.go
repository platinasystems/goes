// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package main

import (
	"io"

	"github.com/platinasystems/go/builtinutils"
	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/coreutils"
	"github.com/platinasystems/go/diagutils"
	"github.com/platinasystems/go/diagutils/dlv"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/fsutils"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/initutils/goesd"
	"github.com/platinasystems/go/kutils"
	"github.com/platinasystems/go/machined"
	"github.com/platinasystems/go/netutils"
	"github.com/platinasystems/go/redisutils"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/bus/pci"
	//"github.com/platinasystems/go/vnet/devices/ethernet/ixge"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	"github.com/platinasystems/go/vnet/unix"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm"

	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	Machine  = "platina-mk1"
	UnixSock = "/run/goes/socks/npu"
)

type platform struct {
	vnet.Package
	*bcm.Platform
}

type parser interface {
	Parse(string) error
}

type Info struct {
	mutex    sync.Mutex
	name     string
	prefixes []string
	attrs    machined.Attrs
}

type AttrInfo struct {
	attr_name string
	attr      interface{}
}

// Would like to do "eth-0-0.speedsetting = 100e9,4" where 4 is number of lanes based off subport 0 here.
var portConfigs = []AttrInfo{
	{"speedsetting", "100e9,4"},
	{"autoneg", false},
	{"loopback", false},
}

var portCounters = []AttrInfo{
	{"Rx_packets", uint64(0)},
	{"Rx_bytes", uint64(0)},
	{"Rx_64_byte_packets", uint64(0)},
	{"Rx_65_to_127_byte_packets", uint64(0)},
	{"Rx_128_to_255_byte_packets", uint64(0)},
	{"Rx_256_to_511_byte_packets", uint64(0)},
	{"Rx_512_to_1023_byte_packets", uint64(0)},
	{"Rx_1024_to_1518_byte_packets", uint64(0)},
	{"Rx_1519_to_1522_byte_vlan_packets", uint64(0)},
	{"Rx_1519_to_2047_byte_packets", uint64(0)},
	{"Rx_2048_to_4096_byte_packets", uint64(0)},
	{"Rx_4096_to_9216_byte_packets", uint64(0)},
	{"Rx_9217_to_16383_byte_packets", uint64(0)},
	{"Rx_good_packets", uint64(0)},
	{"Rx_unicast_packets", uint64(0)},
	{"Rx_multicast_packets", uint64(0)},
	{"Rx_broadcast_packets", uint64(0)},
	{"Rx_crc_error_packets", uint64(0)},
	{"Rx_control_packets", uint64(0)},
	{"Rx_flow_control_packets", uint64(0)},
	{"Rx_pfc_packets", uint64(0)},
	{"Rx_unsupported_opcode_control_packets", uint64(0)},
	{"Rx_unsupported_dst_address_control_packets", uint64(0)},
	{"Rx_src_address_not_unicast_packets", uint64(0)},
	{"Rx_alignment_error_packets", uint64(0)},
	{"Rx_802_3_length_error_packets", uint64(0)},
	{"Rx_code_error_packets", uint64(0)},
	{"Rx_false_carrier_events", uint64(0)},
	{"Rx_oversize_packets", uint64(0)},
	{"Rx_jabber_packets", uint64(0)},
	{"Rx_mtu_check_error_packets", uint64(0)},
	{"Rx_mac_sec_crc_matched_packets", uint64(0)},
	{"Rx_promiscuous_packets", uint64(0)},
	{"Rx_1tag_vlan_packets", uint64(0)},
	{"Rx_2tag_vlan_packets", uint64(0)},
	{"Rx_truncated_packets", uint64(0)},
	{"Rx_xon_to_xoff_priority_0", uint64(0)},
	{"Rx_xon_to_xoff_priority_1", uint64(0)},
	{"Rx_xon_to_xoff_priority_2", uint64(0)},
	{"Rx_xon_to_xoff_priority_3", uint64(0)},
	{"Rx_xon_to_xoff_priority_4", uint64(0)},
	{"Rx_xon_to_xoff_priority_5", uint64(0)},
	{"Rx_xon_to_xoff_priority_6", uint64(0)},
	{"Rx_xon_to_xoff_priority_7", uint64(0)},
	{"Rx_pfc_priority_0", uint64(0)},
	{"Rx_pfc_priority_1", uint64(0)},
	{"Rx_pfc_priority_2", uint64(0)},
	{"Rx_pfc_priority_3", uint64(0)},
	{"Rx_pfc_priority_4", uint64(0)},
	{"Rx_pfc_priority_5", uint64(0)},
	{"Rx_pfc_priority_6", uint64(0)},
	{"Rx_pfc_priority_7", uint64(0)},
	{"Rx_undersize_packets", uint64(0)},
	{"Rx_fragment_packets", uint64(0)},
	{"Rx_eee_lpi_events", uint64(0)},
	{"Rx_eee_lpi_duration", uint64(0)},
	{"Rx_runt_bytes", uint64(0)},
	{"Rx_runt_packets", uint64(0)},
	{"Tx_packets", uint64(0)},
	{"Tx_bytes", uint64(0)},
	{"Tx_64_byte_packets", uint64(0)},
	{"Tx_65_to_127_byte_packets", uint64(0)},
	{"Tx_128_to_255_byte_packets", uint64(0)},
	{"Tx_256_to_511_byte_packets", uint64(0)},
	{"Tx_512_to_1023_byte_packets", uint64(0)},
	{"Tx_1024_to_1518_byte_packets", uint64(0)},
	{"Tx_1519_to_1522_byte_vlan_packets", uint64(0)},
	{"Tx_1519_to_2047_byte_packets", uint64(0)},
	{"Tx_2048_to_4096_byte_packets", uint64(0)},
	{"Tx_4096_to_9216_byte_packets", uint64(0)},
	{"Tx_9217_to_16383_byte_packets", uint64(0)},
	{"Tx_good_packets", uint64(0)},
	{"Tx_unicast_packets", uint64(0)},
	{"Tx_multicast_packets", uint64(0)},
	{"Tx_broadcast_packets", uint64(0)},
	{"Tx_flow_control_packets", uint64(0)},
	{"Tx_pfc_packets", uint64(0)},
	{"Tx_jabber_packets", uint64(0)},
	{"Tx_fcs_errors", uint64(0)},
	{"Tx_control_packets", uint64(0)},
	{"Tx_oversize", uint64(0)},
	{"Tx_single_deferral_packets", uint64(0)},
	{"Tx_multiple_deferral_packets", uint64(0)},
	{"Tx_single_collision_packets", uint64(0)},
	{"Tx_multiple_collision_packets", uint64(0)},
	{"Tx_late_collision_packets", uint64(0)},
	{"Tx_excessive_collision_packets", uint64(0)},
	{"Tx_fragments", uint64(0)},
	{"Tx_system_error_packets", uint64(0)},
	{"Tx_1tag_vlan_packets", uint64(0)},
	{"Tx_2tag_vlan_packets", uint64(0)},
	{"Tx_runt_packets", uint64(0)},
	{"Tx_fifo_underrun_packets", uint64(0)},
	{"Tx_pfc_priority0_packets", uint64(0)},
	{"Tx_pfc_priority1_packets", uint64(0)},
	{"Tx_pfc_priority2_packets", uint64(0)},
	{"Tx_pfc_priority3_packets", uint64(0)},
	{"Tx_pfc_priority4_packets", uint64(0)},
	{"Tx_pfc_priority5_packets", uint64(0)},
	{"Tx_pfc_priority6_packets", uint64(0)},
	{"Tx_pfc_priority7_packets", uint64(0)},
	{"Tx_eee_lpi_events", uint64(0)},
	{"Tx_eee_lpi_duration", uint64(0)},
	{"Tx_total_collisions", uint64(0)},
}

// need to add mmu
var datapathCounters = []AttrInfo{
	{"rx_ip4_l3_drops", uint64(0)},
	{"rx_ip4_l3_packets", uint64(0)},
	{"rx_ip4_header_errors", uint64(0)},
	{"rx_ip4_routed_multicast_packets", uint64(0)},
	{"rx_ip6_l3_drops", uint64(0)},
	{"rx_ip6_l3_packets", uint64(0)},
	{"rx_ip6_header_errors", uint64(0)},
	{"rx_ip6_routed_multicast_packets", uint64(0)},
	{"rx_ibp_discard_cbp_full_drops", uint64(0)},
	{"rx_unicast_packets", uint64(0)},
	{"rx_spanning_tree_state_not_forwarding_drops", uint64(0)},
	{"rx_debug_0", uint64(0)},
	{"rx_debug_1", uint64(0)},
	{"rx_debug_2", uint64(0)},
	{"rx_debug_3", uint64(0)},
	{"rx_debug_4", uint64(0)},
	{"rx_debug_5", uint64(0)},
	{"rx_hi_gig_unknown_hgi_packets", uint64(0)},
	{"rx_hi_gig_control_packets", uint64(0)},
	{"rx_hi_gig_broadcast_packets", uint64(0)},
	{"rx_hi_gig_l2_multicast_packets", uint64(0)},
	{"rx_hi_gig_l3_multicast_packets", uint64(0)},
	{"rx_hi_gig_unknown_opcode_packets", uint64(0)},
	{"rx_debug_6", uint64(0)},
	{"rx_debug_7", uint64(0)},
	{"rx_debug_8", uint64(0)},
	{"rx_trill_packets", uint64(0)},
	{"rx_trill_trill_drops", uint64(0)},
	{"rx_trill_non_trill_drops", uint64(0)},
	{"rx_niv_frame_error_drops", uint64(0)},
	{"rx_niv_forwarding_error_drops", uint64(0)},
	{"rx_niv_frame_vlan_tagged", uint64(0)},
	{"rx_ecn_counter", uint64(0)},
	{"tx_debug_0", uint64(0)},
	{"tx_debug_1", uint64(0)},
	{"tx_debug_2", uint64(0)},
	{"tx_debug_3", uint64(0)},
	{"tx_debug_4", uint64(0)},
	{"tx_debug_5", uint64(0)},
	{"tx_debug_6", uint64(0)},
	{"tx_debug_7", uint64(0)},
	{"tx_debug_8", uint64(0)},
	{"tx_debug_9", uint64(0)},
	{"tx_debug_a", uint64(0)},
	{"tx_debug_b", uint64(0)},
	{"tx_trill_packets", uint64(0)},
	{"tx_trill_access_port_drops", uint64(0)},
	{"tx_trill_non_trill_drops", uint64(0)},
	{"tx_ecn_errors", uint64(0)},
	{"tx_purge_cell_error_drops", uint64(0)},
}

func main() {
	command.Plot(builtinutils.New()...)
	command.Plot(coreutils.New()...)
	command.Plot(dlv.New()...)
	command.Plot(diagutils.New()...)
	command.Plot(fsutils.New()...)
	command.Plot(goesd.New())
	command.Plot(kutils.New()...)
	command.Plot(netutils.New()...)
	command.Plot(redisutils.New()...)
	command.Sort()
	machined.Hook = func() {
		machined.NetLink.Prefixes("lo.", "eth0.")
		machined.InfoProviders = append(machined.InfoProviders, &Info{
			name:     "mk1",
			prefixes: []string{"eth-", "dp-"},
			attrs:    make(machined.Attrs),
		})
	}
	os.Setenv("REDISD_DEVS", "lo eth0")
	err := goes.Main(os.Args...)
	if err != nil && err != io.EOF {
		os.Exit(1)
	}
}

func (p *Info) Main(...string) error {
	machined.Publish("machine", "platina-mk1")
	for _, entry := range []struct{ name, unit string }{
		{"current", "milliamperes"},
		{"fan", "% max speed"},
		{"potential", "volts"},
		{"temperature", "°C"},
	} {
		machined.Publish("unit."+entry.name, entry.unit)
	}

	for port := 0; port < 32; port++ {
		for subport := 0; subport < 4; subport++ {
			// Initially only config subport 0 to match default
			if subport == 0 {
				for i := range portConfigs {
					k := fmt.Sprintf("eth-%02d-%d.%s", port, subport,
						portConfigs[i].attr_name)
					p.attrs[k] = portConfigs[i].attr
					// Publish configuration redis nodes
					machined.Publish(k, fmt.Sprint(p.attrs[k]))
				}
				for i := range portCounters {
					p.attrs[fmt.Sprintf("eth-%02d-%d.%s", port, subport,
						portCounters[i].attr_name)] = &portCounters[i].attr
					// Don't publish in redis until non-zero counters come back from device
				}
			}
		}
	}

	for i := range datapathCounters {
		p.attrs[fmt.Sprintf("dp-%s", datapathCounters[i].attr_name)] = &datapathCounters[i].attr
		// Don't publish in redis until non-zero counters come back from device
	}

	var in parse.Input
	vnetArgsLine := fmt.Sprint("cli { listen { socket ", UnixSock, " no-prompt } }")
	vnetArgs := strings.Split(vnetArgsLine, " ")
	in.Add(vnetArgs[0:]...)
	v := &vnet.Vnet{}

	bcm.Init(v)
	ethernet.Init(v)
	ip4.Init(v)
	ip6.Init(v)
	// Temporarily remove to get goesdeb running vnet
	//ixge.Init(v)
	pci.Init(v)
	pg.Init(v)
	unix.Init(v)

	plat := &platform{}
	v.AddPackage("platform", plat)
	plat.DependsOn("pci-discovery") // after pci discovery

	return v.Run(&in)
}

func (p *Info) Close() error {
	return nil
}

func (p *Info) Del(key string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, found := p.attrs[key]; !found {
		return machined.CantDel(key)
	}
	delete(p.attrs, key)
	machined.Publish("delete", key)
	return nil
}

func (p *Info) Prefixes(prefixes ...string) []string {
	if len(prefixes) > 0 {
		p.prefixes = prefixes
	}
	return p.prefixes
}

// Send message to hw channel
func (p *Info) setHw(key, value string) error {
	// If previous configuration existed on this
	// port, delete and start again.

	// Send new setting to vnetdevices layer via channel

	return nil
}

func (p *Info) settableKey(key string) error {
	var (
		found bool
	)
	keyStr := strings.SplitN(key, ".", 2)
	for i := range portConfigs {
		if portConfigs[i].attr_name == keyStr[1] {
			found = true
			break
		}
	}
	if !found {
		return machined.CantSet(key)
	}
	return nil
}

func (p *Info) Set(key, value string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	a, found := p.attrs[key]
	if !found {
		return machined.CantSet(key)
	}

	// Test if this attribute is settable.
	errPerm := p.settableKey(key)
	if errPerm != nil {
		return errPerm
	}

	// Parse key to find port/subport and attribute
	// and send value down to the driver for validation.
	// If all good i.e. hw has set the value, publish it
	errHw := p.setHw(key, value)
	if errHw != nil {
		return errHw
	}

	switch t := a.(type) {
	case string:
		p.attrs[key] = value
	case int:
		i, err := strconv.ParseInt(value, 0, 0)
		if err != nil {
			return err
		}
		p.attrs[key] = i
	case uint64:
		u64, err := strconv.ParseUint(value, 0, 64)
		if err != nil {
			return err
		}
		p.attrs[key] = u64
	case float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		p.attrs[key] = f
	default:
		if method, found := t.(parser); found {
			if err := method.Parse(value); err != nil {
				return err
			}
		} else {
			return machined.CantSet(key)
		}
	}
	machined.Publish(key, fmt.Sprint(p.attrs[key]))
	return nil
}

func (p *Info) String() string { return p.name }
