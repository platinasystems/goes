// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package m

import (
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/i2c"

	"fmt"
)

type SwitchSelect struct {
	Vnet     *vnet.Vnet
	Switches []Switch
}

func (s *SwitchSelect) Parse(in *parse.Input) {
	s.Switches = GetPlatform(s.Vnet).Switches // for now
}

func (s *SwitchSelect) SelectAll() {
	s.Switches = GetPlatform(s.Vnet).Switches
}

func (s *SwitchSelect) SelectFromInput(in *cli.Input) (err error) {
	s.SelectAll()
	for !in.End() {
		switch {
		case in.Parse("d%ev %v", s):
		default:
			err = cli.ParseError
			return
		}
	}
	return
}

func (sws *SwitchSelect) showSwitches(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	sws.SelectFromInput(in)
	for _, s := range sws.Switches {
		fmt.Fprintln(w, s.String())
	}
	return
}

func (sws *SwitchSelect) showBcmInterrupt(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	sws.SelectFromInput(in)
	for _, s := range sws.Switches {
		s.GetSwitchCommon().Cmic.WriteInterruptSummary(w)
	}
	return
}

func i2cScan(sw Switch) {
	i2cBus := sw.GetSwitchCommon().Cmic.I2c
	for a := 0; a < 128; a++ {
		var d i2c.Data
		n, err := i2cBus.Do(i2c.Write, byte(a), i2c.Quick, 0, &d, 0)
		switch {
		case err == i2c.ErrDeviceNotPresent:
		case err != nil:
			fmt.Printf("%02x: read error: %s\n", a, err)
		default:
			fmt.Printf("%02x: %x\n", a, d[:n])
		}
	}
}

func (sws *SwitchSelect) i2cScanAction(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	sws.SelectFromInput(in)
	for _, s := range sws.Switches {
		i2cScan(s)
	}
	return
}

func cliInit(v *vnet.Vnet) {
	s := &SwitchSelect{Vnet: v}
	cmds := []cli.Command{
		cli.Command{
			Name:   "i2c scan",
			Action: s.i2cScanAction,
		},
		cli.Command{
			Name:   "show fe1 switches",
			Action: s.showSwitches,
		},
		cli.Command{
			Name:   "show fe1 interrupt",
			Action: s.showBcmInterrupt,
		},
	}
	for i := range cmds {
		v.CliAdd(&cmds[i])
	}
}
