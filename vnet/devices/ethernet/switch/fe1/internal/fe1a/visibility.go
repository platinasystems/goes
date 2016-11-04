// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"

	"fmt"
)

type rx_pipe_packet_capture struct {
	ivp  [2]uint64
	isw1 [4]uint64
	isw2 [8]uint64
}

func (ss *switchSelect) show_rx_pipe_visibility_packet_capture(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	if err = ss.SelectFromInput(in); err != nil {
		return
	}
	for _, sw := range ss.Switches {
		t := sw.(*fe1a)
		q := t.getDmaReq()
		for pipe := uint(0); pipe < 4; pipe++ {
			var c rx_pipe_packet_capture
			for i := range c.ivp {
				t.rx_pipe_mems.visibility_packet_capture_buffer_ivp[i].Get(&q.DmaRequest, BlockRxPipe, sbus.Unique(pipe), &c.ivp[i])
			}
			for i := range c.isw1 {
				t.rx_pipe_mems.visibility_packet_capture_buffer_isw1[i].Get(&q.DmaRequest, BlockRxPipe, sbus.Unique(pipe), &c.isw1[i])
			}
			for i := range c.isw2 {
				t.rx_pipe_mems.visibility_packet_capture_buffer_isw2[i].Get(&q.DmaRequest, BlockRxPipe, sbus.Unique(pipe), &c.isw2[i])
			}
			q.Do()
			fmt.Fprintf(w, "pipe %d: %x\n", pipe, &c)
		}
	}
	return
}
