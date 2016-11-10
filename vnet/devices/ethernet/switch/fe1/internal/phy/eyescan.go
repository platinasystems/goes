// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phy

import (
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"

	"fmt"
	"strconv"
	"time"
)

func (ss *switchSelect) showEyeScan(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	var ifs vnet.HwIfChooser
	var port string
	ifs.Init(ss.Vnet)

	ss.SelectAll()
	inc := *in
	inp := &inc
	for !in.End() {
		switch {
		case in.Parse("%v", &ifs):
		case in.Parse("d%*ev %v", ss):
		default:
			err = cli.ParseError
			return
		}
	}

	inp.Parse("%s", &port)
	laneIndex, _ := strconv.Atoi(string(port[len(port)-1]))

	for _, s := range ss.Switches {
		ifs.Foreach(func(v *vnet.Vnet, r vnet.HwInterfacer) {
			var (
				p  m.Porter
				ok bool
			)
			if p, ok = r.(m.Porter); !ok {
				fmt.Fprintln(w, s.String())
				return
			}
			phy := p.GetPhy().(*Tscf)
			phy.printEyeScan(p, uint(laneIndex), w)
		})
	}
	return
}

const (
	y_max = 31
	n_y   = 2*y_max + 1
	x_max = 31
	n_x   = 2*x_max + 1
)

type eyeScan struct {
	bitErrorRate [n_y][n_x]float64
}

func (phy *Tscf) printEyeScan(port m.Porter, lane uint, w cli.Writer) {
	fmt.Fprintf(w, "Starting eye for port %s\n", port.GetPortName())

	result := &eyeScan{}

	if err := phy.doEyeScan(result, port, lane); err != nil {
		fmt.Fprintf(w, "Failed: %s\n", err)
		return
	}

	// Read and display eyescan results.
	for iy := 0; iy < n_y; iy++ {
		y := y_max - iy

		// Print header.
		if y == y_max {
			fmt.Fprintf(w, "Sample of I for bit error rate < 1e-I\n")
			fmt.Fprintf(w, "UI/64\t: -30  -25  -20  -15  -10  -5    0    5    10   15   20   25   30\n")
			fmt.Fprintf(w, "\t: -|----|----|----|----|----|----|----|----|----|----|----|----|-\n")
		}

		// Print row.
		fmt.Fprintf(w, "%4.0f mV\t: ", float64(y*600/127))

		for x := 0; x < n_x; x++ {
			ber := result.bitErrorRate[iy][x]
			var c byte
			switch {
			case ber < 1e-8:
				c = ' '
				if (x%5) == 0 && (y%5) == 0 {
					c = '+'
				} else if (x%5) != 0 && (y%5) == 0 {
					c = '-'
				} else if (x%5) == 0 && (y%5) != 0 {
					c = ':'
				}
			case ber < 1e-7:
				c = '7'
			case ber < 1e-6:
				c = '6'
			case ber < 1e-5:
				c = '5'
			case ber < 1e-4:
				c = '4'
			case ber < 1e-3:
				c = '3'
			case ber < 1e-2:
				c = '2'
			case ber < 1e-1:
				c = '1'
			case ber < 1:
				c = '0'
			}
			fmt.Fprintf(w, "%c", c)
		}
		fmt.Fprintf(w, "\n")
	}

	// Print footer.
	fmt.Fprintf(w, "\t: -|----|----|----|----|----|----|----|----|----|----|----|----|-\n")
	fmt.Fprintf(w, "UI/64\t: -30  -25  -20  -15  -10  -5    0    5    10   15   20   25   30\n")
}

// [7:5] 3 bit fraction
// [4:0] 5 bit exponent
type float8 uint8

func (x float8) bitErrorRate() float64 {
	var i uint32
	if x != 0 {
		fraction := (1 << 3) + x>>5
		exp := x & 0x1f
		if exp < 3 {
			i = uint32(fraction) >> (3 - exp)
		} else {
			i = uint32(fraction) << (exp - 3)
		}
	}
	const c = 1 / 18350080.
	return float64(i) * c
}

func (phy *Tscf) doEyeScan(result *eyeScan, port m.Porter, lane uint) (err error) {
	laneMask := m.LaneMask(1 << lane)
	r := get_tscf_regs()
	q := phy.dmaReq()

	//check PMD is locked status
	if !phy.PmdLocked(laneMask) {
		err = fmt.Errorf("pmd not locked")
		return
	}

	// Start eye scan.
	{
		cmd := uc_cmd{
			command:     uc_cmd_diagnostics,
			sub_command: uc_cmd_diagnostics_start_horizontal_eye_scan,
		}
		err = cmd.do(q, laneMask, &r.uc_cmd)
		if err != nil {
			return
		}
	}

	uc_mem := phy.get_uc_mem()

	// Read and display eyescan results.
	for iy := 0; iy < n_y; iy++ {
		for x := 0; x < n_x; x += 2 {
			// Wait for lane diag status to become ready.
			{
				const (
					diag_status_done = 1 << 15
				)
				start := time.Now()
				for {
					v := uc_mem.lanes[lane].diag_status.Get(q, laneMask)
					// Done with scan or we have at least 2 samples in the uc buffer?
					if v&diag_status_done != 0 || (v&0xff) > 2 {
						break
					}
					if time.Since(start) > 100*time.Millisecond {
						err = fmt.Errorf("timeout diag status: 0x%x", v)
						return
					}
					time.Sleep(500 * time.Microsecond)
				}
			}

			// Read uc diag word.
			var v uint16
			cmd := uc_cmd{
				command: uc_cmd_read_diagnostic_data_word,
				out:     &v,
			}
			err = cmd.do(q, laneMask, &r.uc_cmd)
			if err != nil {
				return
			}

			// Each diag word contains 2 8 bit floating point samples.
			v0, v1 := float8(v>>8), float8(v)
			result.bitErrorRate[iy][x+0] = v0.bitErrorRate()
			if x+1 < n_x {
				result.bitErrorRate[iy][x+1] = v1.bitErrorRate()
			}
		}
	}

	// Stop eye-scan.
	{
		cmd := uc_cmd{
			command:     uc_cmd_diagnostics,
			sub_command: uc_cmd_diagnostics_disable,
		}
		err = cmd.do(q, laneMask, &r.uc_cmd)
		if err != nil {
			return
		}
	}

	return
}
