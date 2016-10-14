// +build debug

package elib

import (
	"github.com/platinasystems/go/elib/cpu"

	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

type uiHash struct {
	Hash
	pairs uiPairVec
}

type uiKey uint64
type uiValue uint64
type uiPair struct {
	key   uiKey
	value uiValue
}

func (p *uiPair) Equal(q *uiPair) bool { return p.key == q.key && p.value == q.value }

//go:generate gentemplate -d Package=elib -id uiPair -d VecType=uiPairVec -d Type=uiPair -tags debug vec.tmpl

func (k *uiKey) HashKey(s *HashState)               { s.HashUint64(uint64(*k), 0, 0, 0) }
func (k *uiKey) HashKeyEqual(h Hasher, i uint) bool { return *k == h.(*uiHash).pairs[i].key }
func (h *uiHash) HashIndex(s *HashState, i uint)    { h.pairs[i].key.HashKey(s) }
func (h *uiHash) HashResize(newCap uint, rs []HashResizeCopy) {
	src, dst := h.pairs, make([]uiPair, newCap)
	for i := range rs {
		dst[rs[i].Dst] = src[rs[i].Src]
	}
	h.pairs = dst
}

type testHash struct {
	uiHash uiHash

	pairs    uiPairVec
	inserted Bitmap

	// Number of iterations to run
	iterations Count

	// Validate/print every so many iterations (zero means never).
	validateEvery Count
	printEvery    Count

	// Seed to make randomness deterministic.  0 means choose seed.
	seed int64

	nKeys Count

	verbose  int
	testTime bool

	profile string
}

func HashTest() {
	t := testHash{
		iterations: 10,
		nKeys:      10,
		verbose:    1,
	}
	flag.Var(&t.iterations, "iter", "Number of iterations")
	flag.Var(&t.validateEvery, "valid", "Number of iterations per validate")
	flag.Var(&t.printEvery, "print", "Number of iterations per print")
	flag.Int64Var(&t.seed, "seed", 0, "Seed for random number generator")
	flag.Var(&t.nKeys, "keys", "Number of random keys")
	flag.IntVar(&t.verbose, "verbose", 0, "Be verbose")
	flag.BoolVar(&t.testTime, "time", false, "Time hash functions")
	flag.StringVar(&t.profile, "profile", "", "Write CPU profile to file")
	flag.Parse()

	err := runHashTest(&t)
	if err != nil {
		panic(err)
	}
}

func (t *testHash) doValidate() (err error) {
	h := &t.uiHash
	for pi := uint(0); pi < t.pairs.Len(); pi++ {
		p := &t.pairs[pi]
		i, ok := h.Get(&p.key)
		if got, want := ok, t.inserted.Get(pi); got != want {
			err = fmt.Errorf("get ok %v != inserted %v", got, want)
			return
		}
		if ok && !p.Equal(&h.pairs[i]) {
			err = fmt.Errorf("get index got %d != want %d", i, pi)
			return
		}
	}
	return
}

func (t *testHash) validate(h *Hash, iter int) (err error) {
	if t.validateEvery != 0 && iter%int(t.validateEvery) == 0 {
		if err = t.doValidate(); err != nil {
			if t.verbose != 0 {
				fmt.Printf("iter %d: %s\n", iter, err)
			}
			return
		}
	}
	if t.printEvery != 0 && iter%int(t.printEvery) == 0 {
		fmt.Printf("%10g iterations: %s\n", float64(iter), h)
	}
	return
}

func runHashTest(t *testHash) (err error) {
	if t.seed == 0 {
		t.seed = int64(time.Now().Nanosecond())
	}

	rand.Seed(t.seed)
	if t.verbose != 0 {
		fmt.Printf("%#v\n", t)
	}

	h := &t.uiHash
	t.pairs.Resize(uint(t.nKeys))
	log2n := Word(t.nKeys).MaxLog2()
	for i := range t.pairs {
		t.pairs[i].key = uiKey((uint64(rand.Int63()) << log2n) + uint64(i))
		t.pairs[i].value = uiValue(rand.Int63())
	}

	if t.profile != "" {
		var f *os.File
		f, err = os.Create(t.profile)
		if err != nil {
			return
		}
		pprof.StartCPUProfile(f)
		defer func() { pprof.StopCPUProfile() }()
	}

	if t.testTime {
		t.timeHash()
		return
	}

	h.Hasher = h
	zero := uiPair{}
	start := time.Now()
	var iter int
	for ; iter < int(t.iterations); iter++ {
		pi := uint(rand.Intn(int(t.nKeys)))
		p := &t.pairs[pi]
		var was bool
		if t.inserted, was = t.inserted.Invert2(pi); !was {
			i, exists := h.Set(&p.key)
			if exists {
				panic("exists")
			}
			h.pairs[i] = *p
		} else {
			i, ok := h.Unset(&p.key)
			if !ok {
				panic("unset")
			}
			h.pairs[i] = zero
		}

		err = t.validate(&h.Hash, iter)
		if err != nil {
			return
		}
	}
	if t.verbose != 0 {
		fmt.Printf("%d iterations: %s\n", iter, h)
	}
	for _ = range h.pairs {
		err = t.validate(&h.Hash, iter)
		if err != nil {
			return
		}
		iter++
	}
	dt := time.Since(start)
	fmt.Printf("%d iterations: %e iter/sec %s\n", iter, float64(iter)/dt.Seconds(), h)
	return
}

func (t *testHash) timeHash() {
	h := &t.uiHash
	h.Init(h, uint(2*t.nKeys))
	var tm struct {
		set, get, unset, delete cpu.Timing
	}

	{
		niter := uint(len(t.pairs))
		if niter < uint(t.iterations) {
			niter = uint(t.iterations)
		}
		tm.set[0] = cpu.TimeNow()
		pi := 0
		for iter := uint(0); iter < niter; iter++ {
			i, _ := h.Set(&t.pairs[pi].key)
			h.pairs[i] = t.pairs[pi]
			pi++
			if pi >= len(t.pairs) {
				pi = 0
			}
		}
		tm.set[1] = cpu.TimeNow()
		fmt.Printf("set: %.2f clocks/operation, %e per sec\n", tm.set.ClocksPer(niter), tm.set.PerSecond(niter))
	}

	{
		tm.get[0] = cpu.TimeNow()
		pi := 0
		niter := uint(t.iterations)
		for iter := uint(0); iter < niter; iter++ {
			i, _ := h.Get(&t.pairs[pi].key)
			h.pairs[i].value++
		}
		tm.get[1] = cpu.TimeNow()
		fmt.Printf("get: %.2f clocks/operation, %e per sec\n", tm.get.ClocksPer(niter), tm.set.PerSecond(niter))
	}

	{
		tm.unset[0] = cpu.TimeNow()
		niter := uint(len(t.pairs))
		for pi := range t.pairs {
			i, _ := h.Unset(&t.pairs[pi].key)
			h.pairs[i].value = 0
		}
		tm.unset[1] = cpu.TimeNow()
		fmt.Printf("unset: %.2f clocks/operation, %e per sec\n", tm.unset.ClocksPer(niter), tm.unset.PerSecond(niter))
	}
}
