package main

import (
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/elog/elogview"

	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"
)

type ev struct {
	i     uint32
	color color
}

type color uint32

var colorNames = [...]string{
	0: "red",
	1: "green",
	2: "blue",
	3: "yellow",
}

func (c color) String() string                             { return colorNames[c] }
func (e *ev) SetData(x *elog.Context, p elog.Pointer)      { *(*ev)(p) = *e }
func (e *ev) Format(x *elog.Context, f elog.Format) string { return f("%s %d", e.color.String(), e.i) }

func main() {
	var (
		n_events     uint
		delay        float64
		random_delay bool
		save, load   string
		useFmt       bool
	)
	flag.Float64Var(&delay, "delay", 0, "delay in seconds between events or max delay for random delays.")
	flag.UintVar(&n_events, "events", 10, "number of test events to add")
	flag.BoolVar(&random_delay, "random", false, "randomize delays")
	flag.StringVar(&load, "load", "", "load log from file")
	flag.StringVar(&save, "save", "", "save log to file")
	flag.BoolVar(&useFmt, "fmt", false, "use elog.F* formatted functions")
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
		var ms [2]runtime.MemStats
		runtime.ReadMemStats(&ms[0])
		var e ev
		for i := uint64(0); i < uint64(n_events); i++ {
			if useFmt {
				switch i % 4 {
				case 0:
					elog.FUint("red wjof owfj owjf wofjwf %d", i)
				case 1:
					elog.FUint("green %d", i)
				case 2:
					elog.FUint("blue %d", i)
				case 3:
					elog.FUint("yellow %d", i)
				}
			} else {
				e.color = color(i % 4)
				e.i = uint32(i)
				switch i % 4 {
				case 0:
					elog.Add(&e)
				case 1:
					elog.Add(&e)
				case 2:
					elog.Add(&e)
				case 3:
					elog.Add(&e)
				}
			}
			if delay > 0 {
				d := delay
				if random_delay {
					d = rand.Float64() * delay
				}
				time.Sleep(time.Duration(1e9 * d))
			}
		}
		runtime.ReadMemStats(&ms[1])
		fmt.Printf("mallocs %d gcs %d\n", ms[1].Mallocs-ms[0].Mallocs, ms[1].NumGC-ms[0].NumGC)
		v = elog.NewView()
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
