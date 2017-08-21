package main

import (
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/elog/elogview"

	"flag"
	"math/rand"
	"os"
	"time"
)

func main() {
	var (
		n_events     uint
		delay        float64
		random_delay bool
		save, load   string
	)
	flag.Float64Var(&delay, "delay", 0, "delay in seconds between events or max delay for random delays.")
	flag.UintVar(&n_events, "events", 10, "number of test events to add")
	flag.BoolVar(&random_delay, "random", false, "randomize delays")
	flag.StringVar(&load, "load", "", "load log from file")
	flag.StringVar(&save, "save", "", "save log to file")
	flag.Parse()

	var v *elog.View

	if load != "" {
		if f, err := os.OpenFile(load, os.O_RDONLY, 0); err != nil {
			panic(err)
		} else {
			defer f.Close()
			var r elog.View
			if err = r.Restore(f); err != nil {
				panic(err)
			}
			v = &r
		}
	} else {
		elog.DefaultBuffer.Resize(n_events)
		elog.Enable(true)
		for i := uint(0); i < n_events; i++ {
			switch i % 4 {
			case 0:
				elog.F("red wjof owfj owjf wofjwf %d", i)
			case 1:
				elog.F("green %d", i)
			case 2:
				elog.F("blue %d", i)
			case 3:
				elog.F("yellow %d", i)
			}
			d := delay
			if random_delay {
				d = rand.Float64() * delay
			}
			time.Sleep(time.Duration(1e9 * d))
			v = elog.NewView()
		}
	}
	if save != "" {
		if f, err := os.OpenFile(save, os.O_CREATE|os.O_RDWR, 0666); err != nil {
			panic(err)
		} else {
			defer f.Close()
			if err = v.Save(f); err != nil {
				panic(err)
			}
		}
	}

	cf := elogview.Config{
		Width:              1200,
		Height:             750,
		EnableKeyboardQuit: true,
	}
	elogview.View(v, cf)
}
