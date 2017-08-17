// go build -o ~/y -tags "elog gtk_3_16" -gcflags "-N -l" github.com/platinasystems/go/wip/elogview

package main

import (
	"github.com/platinasystems/go/elib/elog"

	"flag"
	"math/rand"
	"sync"
	"time"
)

func main() {
	var (
		n_events     uint
		delay        float64
		random_delay bool
	)
	flag.UintVar(&n_events, "events", 10, "number of test events to add")
	flag.Float64Var(&delay, "delay", 0, "delay in seconds between events or max delay for random delays.")
	flag.BoolVar(&random_delay, "random", false, "randomize delays")
	flag.Parse()

	elog.DefaultBuffer.Resize(n_events)
	elog.Enable(true)
	for i := uint(0); i < n_events; i++ {
		switch i % 4 {
		case 0:
			elog.GenEventf("red wjof owfj owjf wofjwf %d", i)
		case 1:
			elog.GenEventf("green %d", i)
		case 2:
			elog.GenEventf("blue %d", i)
		case 3:
			elog.GenEventf("yellow %d", i)
		}
		d := delay
		if random_delay {
			d = rand.Float64() * delay
		}
		time.Sleep(time.Duration(1e9 * d))
	}

	v := elog.NewView()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		elog_viewer(v, 1200, 750)
		wg.Done()
	}()
	wg.Wait()
}
