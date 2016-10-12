package ixge

import (
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/go/vnet"

	"fmt"
)

type counter struct {
	offset   uint32
	is_64bit bool
	name     string
}

const (
	rx_packets = iota
	rx_bytes
	rx_multicast_packets
	rx_broadcast_packets
	rx_64_byte_packets
	rx_65_127_byte_packets
	rx_128_255_byte_packets
	rx_256_511_byte_packets
	rx_512_1023_byte_packets
	rx_gt_1023_byte_packets
	rx_crc_errors
	rx_ip4_tcp_udp_checksum_errors
	rx_illegal_symbol_errors
	rx_error_symbol_errors
	rx_mac_local_faults
	rx_mac_remote_faults
	rx_length_errors
	rx_xons
	rx_xoffs
	rx_undersize_packets
	rx_fragments
	rx_oversize_packets
	rx_jabbers
	rx_management_packets
	rx_management_drops
	rx_missed_packets_pool_0
	rx_pre_filter_good_packets
	rx_pre_filter_good_bytes
	rx_post_filter_good_packets
	rx_post_filter_good_bytes
	rx_dma_good_packets
	rx_dma_good_bytes
	rx_dma_duplicated_good_packets
	rx_dma_duplicated_good_bytes
	rx_dma_good_loopback_packets
	rx_dma_good_loopback_bytes
	rx_dma_good_duplicated_loopback_packets
	rx_dma_good_duplicated_loopback_bytes
	tx_packets
	tx_good_packets
	tx_good_bytes
	tx_multicast_packets
	tx_broadcast_packets
	tx_dma_good_packets
	tx_dma_good_bytes
	tx_64_byte_packets
	tx_65_127_byte_packets
	tx_128_255_byte_packets
	tx_256_511_byte_packets
	tx_512_1023_byte_packets
	tx_gt_1023_byte_packets
	tx_undersize_drops
	n_counters
)

var counters = [n_counters]counter{
	rx_packets:                              counter{offset: 0x40d0, name: "rx packets"},
	rx_bytes:                                counter{offset: 0x40c0, name: "rx bytes", is_64bit: true},
	rx_multicast_packets:                    counter{offset: 0x407c, name: "rx multicast packets"},
	rx_broadcast_packets:                    counter{offset: 0x4078, name: "rx broadcast packets"},
	rx_64_byte_packets:                      counter{offset: 0x405c, name: "rx 64 byte packets"},
	rx_65_127_byte_packets:                  counter{offset: 0x4060, name: "rx 65 to 127 byte packets"},
	rx_128_255_byte_packets:                 counter{offset: 0x4064, name: "rx 128 to 255 byte packets"},
	rx_256_511_byte_packets:                 counter{offset: 0x4068, name: "rx 256 to 511 byte packets"},
	rx_512_1023_byte_packets:                counter{offset: 0x406c, name: "rx 512 to 1023 byte packets"},
	rx_gt_1023_byte_packets:                 counter{offset: 0x4070, name: "rx 1024 or greater byte packets"},
	rx_crc_errors:                           counter{offset: 0x4000, name: "rx crc errors"},
	rx_ip4_tcp_udp_checksum_errors:          counter{offset: 0x4120, name: "rx ip4/tcp/udp checksum errors"},
	rx_illegal_symbol_errors:                counter{offset: 0x4004, name: "rx illegal symbol errors"},
	rx_error_symbol_errors:                  counter{offset: 0x4008, name: "rx error symbol errors"},
	rx_mac_local_faults:                     counter{offset: 0x4034, name: "rx mac local faults"},
	rx_mac_remote_faults:                    counter{offset: 0x4038, name: "rx mac remote faults"},
	rx_length_errors:                        counter{offset: 0x4040, name: "rx length errors"},
	rx_xons:                                 counter{offset: 0x41a4, name: "rx xons"},
	rx_xoffs:                                counter{offset: 0x41a8, name: "rx xoffs"},
	rx_undersize_packets:                    counter{offset: 0x40a4, name: "rx undersize packets"},
	rx_fragments:                            counter{offset: 0x40a8, name: "rx fragments"},
	rx_oversize_packets:                     counter{offset: 0x40ac, name: "rx oversize packets"},
	rx_jabbers:                              counter{offset: 0x40b0, name: "rx jabbers"},
	rx_management_packets:                   counter{offset: 0x40b4, name: "rx management packets"},
	rx_management_drops:                     counter{offset: 0x40b8, name: "rx management drops"},
	rx_missed_packets_pool_0:                counter{offset: 0x3fa0, name: "rx missed packets pool 0"},
	rx_pre_filter_good_packets:              counter{offset: 0x41b0, name: "rx pre-filter good packets"},
	rx_pre_filter_good_bytes:                counter{offset: 0x41b4, name: "rx pre-filter good bytes", is_64bit: true},
	rx_post_filter_good_packets:             counter{offset: 0x4074, name: "rx post-filter packets"},
	rx_post_filter_good_bytes:               counter{offset: 0x4088, name: "rx post-filter bytes", is_64bit: true},
	rx_dma_good_packets:                     counter{offset: 0x2f50, name: "rx dma good packets"},
	rx_dma_good_bytes:                       counter{offset: 0x2f54, name: "rx dma good bytes", is_64bit: true},
	rx_dma_duplicated_good_packets:          counter{offset: 0x2f5c, name: "rx dma duplicated good packets"},
	rx_dma_duplicated_good_bytes:            counter{offset: 0x2f60, name: "rx dma duplicated good bytes", is_64bit: true},
	rx_dma_good_loopback_packets:            counter{offset: 0x2f68, name: "rx dma good loopback packets"},
	rx_dma_good_loopback_bytes:              counter{offset: 0x2f6c, name: "rx dma good loopback bytes", is_64bit: true},
	rx_dma_good_duplicated_loopback_packets: counter{offset: 0x2f74, name: "rx dma good duplicated loopback packets"},
	rx_dma_good_duplicated_loopback_bytes:   counter{offset: 0x2f78, name: "rx dma good duplicated loopback bytes", is_64bit: true},
	tx_packets:                              counter{offset: 0x40d4, name: "tx packets"},
	tx_good_packets:                         counter{offset: 0x4080, name: "tx good packets"},
	tx_good_bytes:                           counter{offset: 0x4090, name: "tx good bytes", is_64bit: true},
	tx_multicast_packets:                    counter{offset: 0x40f0, name: "tx multicast packets"},
	tx_broadcast_packets:                    counter{offset: 0x40f4, name: "tx broadcast packets"},
	tx_dma_good_packets:                     counter{offset: 0x87a0, name: "tx dma good packets"},
	tx_dma_good_bytes:                       counter{offset: 0x87a4, name: "tx dma good bytes", is_64bit: true},
	tx_64_byte_packets:                      counter{offset: 0x40d8, name: "tx 64 byte packets"},
	tx_65_127_byte_packets:                  counter{offset: 0x40dc, name: "tx 65 to 127 byte packets"},
	tx_128_255_byte_packets:                 counter{offset: 0x40e0, name: "tx 128 to 255 byte packets"},
	tx_256_511_byte_packets:                 counter{offset: 0x40e4, name: "tx 256 to 511 byte packets"},
	tx_512_1023_byte_packets:                counter{offset: 0x40e8, name: "tx 512 to 1023 byte packets"},
	tx_gt_1023_byte_packets:                 counter{offset: 0x40ec, name: "tx 1024 or greater byte packets"},
	tx_undersize_drops:                      counter{offset: 0x4010, name: "tx undersize drops"},
}

func (c *counter) get(d *dev) (v uint64) {
	o := uint(c.offset)
	if c.is_64bit {
		v = hw.LoadUint64(d.addr_for_offset64(o))
	} else {
		v = uint64(hw.LoadUint32(d.addr_for_offset32(o)))
	}
	return
}

func (d *dev) foreach_counter(only_64_bit bool, fn func(i uint, v uint64)) {
	for i := range counters {
		c := &counters[i]
		if only_64_bit && !c.is_64bit {
			continue
		}
		// All counters are clear on read; so always add to previous value.
		v := c.get(d)
		if fn != nil {
			fn(uint(i), v)
		}
	}
}

func (d *dev) clear_counters() {
	d.foreach_counter(false, nil)
}

func (d *dev) counter_init() {
	// Clear anything left over from previous runs.
	d.clear_counters()
	d.AddTimedEvent(&counter_update_event{dev: d}, counter_update_interval)
}

func (d *dev) GetHwInterfaceCounterNames() (n vnet.InterfaceCounterNames) {
	// Initialize counters names on first call.
	cn := &d.m.InterfaceCounterNames
	if len(cn.Single) == 0 {
		for i := range counters {
			cn.Single = append(cn.Single, counters[i].name)
		}
	}
	n = *cn
	return
}

func (d *dev) GetHwInterfaceCounterValues(th *vnet.InterfaceThread) {
	hi := d.Hi()
	d.foreach_counter(false, func(i uint, v uint64) {
		vnet.HwIfCounterKind(i).Add64(th, hi, v)
	})
}

type counter_update_event struct {
	vnet.Event
	dev      *dev
	sequence uint
}

func (e *counter_update_event) String() (s string) {
	s = fmt.Sprintf("%s counter update sequence %d: ", e.dev.Name(), e.sequence)
	if e.only_64_bit() {
		s += "64-bit"
	} else {
		s += "32-bit and 64-bit"
	}
	return
}

// We have 32 bit packets counters, 36 bit byte counters.
// Worst case, byte counters may overflow in around a minute at 10G;
// packet counters may overflow in around 5 minutes.
const counter_update_interval = 45

func (e *counter_update_event) only_64_bit() bool {
	return e.sequence%5 != 4 // 32 bit counters 5 times less often
}

func (e *counter_update_event) EventAction() {
	d := e.dev
	hi := d.Hi()
	th := d.m.Vnet.GetIfThread(0)
	only_64_bit := e.only_64_bit()
	d.foreach_counter(only_64_bit, func(i uint, v uint64) {
		vnet.HwIfCounterKind(i).Add64(th, hi, v)
	})
	d.AddTimedEvent(e, counter_update_interval)
	e.sequence++
}
