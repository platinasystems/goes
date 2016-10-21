// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ixge

import (
	"github.com/platinasystems/go/elib/cli"

	"fmt"
)

func (q *rx_dma_queue) dump_ring(w cli.Writer) {
	for i := range q.rx_desc {
		fmt.Fprintf(w, "%03d 0x%04x: %s\n", i, i, &q.rx_desc[i])
	}
}

func (q *tx_dma_queue) dump_ring(w cli.Writer) {
	for i := range q.tx_desc {
		fmt.Fprintf(w, "%03d 0x%04x: %s\n", i, i, &q.tx_desc[i])
	}
}

func (m *main) showDevs(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	for _, dr := range m.devs {
		d := dr.get()

		var v [4]reg
		v[0] = d.regs.interrupt.status_write_1_to_clear.get(d)
		v[1] = d.regs.tx_dma_control.get(d)
		v[2] = d.regs.rx_enable.get(d)
		v[3] = d.regs.xge_mac.mac_control.get(d)
		fmt.Fprintf(w, "%s: link %v, %x\n", d.Hi().Name(m.Vnet), d.get_link_state(), v)
		for i := range d.tx_queues {
			q := &d.tx_queues[i]
			dr := q.get_regs()
			fmt.Fprintf(w, "txq %d: head %d tail %d\n", i, dr.head_index.get(d), dr.tail_index.get(d))
			v[0] = dr.descriptor_address[0].get(d)
			v[1] = dr.descriptor_address[1].get(d)
			v[2] = dr.n_descriptor_bytes.get(d)
			v[3] = dr.control.get(d)
			fmt.Fprintf(w, "%x\n", v)

			if false {
				q.dump_ring(w)
			}
		}

		for i := range d.rx_queues {
			q := &d.rx_queues[i]
			dr := q.get_regs()
			fmt.Fprintf(w, "rxq %d: head %d tail %d\n", i, dr.head_index.get(d), dr.tail_index.get(d))
			v[0] = dr.descriptor_address[0].get(d)
			v[1] = dr.descriptor_address[1].get(d)
			v[2] = dr.n_descriptor_bytes.get(d)
			v[3] = dr.control.get(d)
			fmt.Fprintf(w, "%x\n", v)

			if false {
				q.dump_ring(w)
			}
		}
	}
	return
}

func (m *main) cliInit() {
	v := m.Vnet
	cmds := [...]cli.Command{
		cli.Command{
			Name:      "show ixge",
			ShortHelp: "show Intel 10G interfaces",
			Action:    m.showDevs,
		},
	}
	for i := range cmds {
		v.CliAdd(&cmds[i])
	}
}
