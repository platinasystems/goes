// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/elib/cli"

	"fmt"
)

type tmon_main struct{ *fe1a }

func (ss *switchSelect) show_temp(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	reset := uint(0)
	ss.SelectAll()
	for !in.End() {
		switch {
		case in.Parse("r%*eset %d", &reset):
		case in.Parse("d%*ev", ss):
		default:
			err = cli.ParseError
			return
		}
	}

	for _, sw := range ss.Switches {
		t := sw.(*fe1a)

		if reset == 1 {
			t.resetTemp()
		}

		var results tMon
		for j := range results.Temp {
			results.Temp[j].get(t, j)
			fmt.Fprintf(w, "%d: %+v\n", j, &results.Temp[j])
		}
	}
	return
}

type Temperature struct {
	current, max, min float64
}

type tMon struct {
	Temp [8]Temperature
}

//Read temp sensors
func (m *Temperature) get(t *fe1a, i int) {
	q := t.getDmaReq()
	data := t.top_controller.temperature_sensor.current[i].getDo(q)
	m.current = float64((4100400 - (4870 * (data & 0x3FF))) / 10000.)
	m.max = float64((4100400 - (4870 * ((data >> 12) & 0x3FF))) / 10000.)
	m.min = float64((4100400 - (4870 * ((data >> 22) & 0x3FF))) / 10000.)
	return
}

//Reset min/max tracking values
func (t *fe1a) resetTemp() {
	q := t.getDmaReq()
	const tmon_reset = (1 << 18) | (1 << 19)
	v := t.top_controller.soft_reset[1].getDo(q)
	t.top_controller.soft_reset[1].set(q, v&^tmon_reset)
	t.top_controller.soft_reset[1].set(q, v|tmon_reset)
	q.Do()
	return
}

//Initialize temperature sensors
func (t *fe1a) tmon_init() {
	// read once to clear garbage data
	q := t.getDmaReq()
	var tmon tMon
	for i := range tmon.Temp {
		tmon.Temp[i].get(t, i)
	}

	const BG_ADJF = 0x7
	v := t.top_controller.temperature_sensor.control[0].getDo(q)
	// set BG_ADJF to 0
	t.top_controller.temperature_sensor.control[0].set(q, v&^BG_ADJF)
	q.Do()

	//reset min max tracking
	t.resetTemp()

	//setup max temp interrupt threshold to 125°C on all sensors
	const maxDegC = 125
	const maxTempThresh = (410040 - (maxDegC * 1000)) / 478
	for i := range t.top_controller.temperature_sensor_interrupt.thresholds {
		t.top_controller.temperature_sensor_interrupt.thresholds[i].set(q, (maxTempThresh<<10)|0x3ff)
		q.Do()
		// sensor 8 is unused
		if i == 7 {
			break
		}
	}

	// enable max temp interrupt for sensor 6
	t.top_controller.temperature_sensor_interrupt.enable.set(q, 0x00002000)
	q.Do()

	return
}
