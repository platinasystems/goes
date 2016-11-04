// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
)

type switchSelect struct{ m.SwitchSelect }

func (t *fe1a) cliInit() {
	ss := &switchSelect{}
	ss.Vnet = t.Vnet
	cmds := []cli.Command{
		cli.Command{
			Name:      "show fe1 debug-events",
			ShortHelp: "show fe1 rx/tx pipe debug events",
			Action:    ss.show_rx_tx_pipe_debug_events,
		},
		cli.Command{
			Name:      "show fe1 visibility",
			ShortHelp: "show rx pipe visibility packet capture",
			Action:    ss.show_rx_pipe_visibility_packet_capture,
		},
		cli.Command{
			Name:      "show fe1 temperature",
			ShortHelp: "show fe1 temperature data",
			Action:    ss.show_temp,
		},
	}
	for i := range cmds {
		ss.Vnet.CliAdd(&cmds[i])
	}
}
