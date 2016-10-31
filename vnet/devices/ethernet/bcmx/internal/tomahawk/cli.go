// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tomahawk

import (
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
)

type switchSelect struct{ m.SwitchSelect }

func (t *tomahawk) cliInit() {
	ss := &switchSelect{}
	ss.Vnet = t.Vnet
	cmds := []cli.Command{
		cli.Command{
			Name:      "show bcm debug-events",
			ShortHelp: "show broadcom rx/tx pipe debug events",
			Action:    ss.show_rx_tx_pipe_debug_events,
		},
		cli.Command{
			Name:      "show bcm visibility",
			ShortHelp: "show rx pipe visibility packet capture",
			Action:    ss.show_rx_pipe_visibility_packet_capture,
		},
		cli.Command{
			Name:      "show bcm temperature",
			ShortHelp: "show tomahawk temperature data",
			Action:    ss.show_bcm_temp,
		},
	}
	for i := range cmds {
		ss.Vnet.CliAdd(&cmds[i])
	}
}
