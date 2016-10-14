// +build debug

package elib

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

func (f *FibHeap) validateNode(xi Index) (err error) {
	x := &f.nodes[xi]
	nSub := uint16(0)
	if x.sub != MaxIndex {
		subi := x.sub
		for {
			sub := &f.nodes[subi]

			if sub.sup != xi {
				err = fmt.Errorf("node.sub.sup %d != node %d", sub.sup, xi)
				return
			}

			n := &f.nodes[sub.next]
			p := &f.nodes[sub.prev]
			if n.prev != subi {
				err = fmt.Errorf("next.prev %d != node %d", n.prev, subi)
				return
			}
			if p.next != subi {
				err = fmt.Errorf("prev.next %d != node %d", p.next, subi)
				return
			}

			err = f.validateNode(subi)
			if err != nil {
				return
			}

			nSub++
			subi = sub.next
			if subi == x.sub {
				break
			}
		}

		if nSub != x.nSub {
			err = fmt.Errorf("n children %d != %d", nSub, x.nSub)
			return
		}
	}
	return
}

func (f *FibHeap) validate() (err error) {
	for ri := f.root.next; ri != fibRootIndex; {
		r := &f.nodes[ri]
		n := f.node(r.next)
		if n.prev != ri {
			err = fmt.Errorf("root next.prev %d != %d", n.prev, ri)
			return
		}
		if n.sup != MaxIndex {
			err = fmt.Errorf("root sup not empty")
			return
		}
		err = f.validateNode(ri)
		if err != nil {
			return
		}
		ri = r.next
	}

	return
}

type testFibHeap struct {
	// Number of iterations to run
	iterations Count

	// Validate/print every so many iterations (zero means never).
	validateEvery Count
	printEvery    Count

	// Seed to make randomness deterministic.  0 means choose seed.
	seed int64

	// Number of objects to create.
	nObjects Count

	verbose int

	profile string
}

type fibHeapTestObj []int64

func (data fibHeapTestObj) Compare(i, j int) int {
	return int(data[i] - data[j])
}

func FibHeapTest() {
	t := testFibHeap{
		iterations: 10,
		nObjects:   10,
		verbose:    1,
	}
	flag.Var(&t.iterations, "iter", "Number of iterations")
	flag.Var(&t.validateEvery, "valid", "Number of iterations per validate")
	flag.Var(&t.printEvery, "print", "Number of iterations per print")
	flag.Int64Var(&t.seed, "seed", 0, "Seed for random number generator")
	flag.Var(&t.nObjects, "objects", "Number of random objects")
	flag.IntVar(&t.verbose, "verbose", 0, "Be verbose")
	flag.StringVar(&t.profile, "profile", "", "Write CPU profile to file")
	flag.Parse()

	err := runFibHeapTest(&t)
	if err != nil {
		panic(err)
	}
}

func runFibHeapTest(t *testFibHeap) (err error) {
	var f FibHeap

	if t.seed == 0 {
		t.seed = int64(time.Now().Nanosecond())
	}

	rand.Seed(t.seed)
	if t.verbose != 0 {
		fmt.Printf("%#v\n", t)
	}
	objs := fibHeapTestObj(make([]int64, t.nObjects))

	var iter int

	validate := func() (err error) {
		i, _ := f.Min(objs)
		fmin := Index(i)
		omin := MaxIndex
		if t.validateEvery != 0 && iter%int(t.validateEvery) == 0 {
			for i := range objs {
				if objs[i] == 0 {
					continue
				}
				if omin == MaxIndex {
					omin = Index(i)
				} else if objs[i] < objs[omin] {
					omin = Index(i)
				}
			}
		} else {
			omin = fmin
		}
		if omin != fmin {
			err = fmt.Errorf("iter %d: min %d != %d", iter, omin, fmin)
			return
		}

		if t.validateEvery != 0 && iter%int(t.validateEvery) == 0 {
			if err = f.validate(); err != nil {
				if t.verbose != 0 {
					fmt.Printf("iter %d: %s\n%+v\n", iter, err, f)
				}
				return
			}
		}
		if t.printEvery != 0 && iter%int(t.printEvery) == 0 {
			fmt.Printf("%10g iterations: %s\n", float64(iter), &f)
		}
		return
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

	for iter = 0; iter < int(t.iterations); iter++ {
		x := uint(rand.Int() % len(objs))
		if objs[x] == 0 {
			objs[x] = 1 + rand.Int63()
			f.Add(x)
		} else {
			if rand.Int()%10 < 5 {
				objs[x] = 1 + rand.Int63()
				f.Update(x)
			} else {
				objs[x] = 0
				f.Del(x)
			}
		}
		err = validate()
		if err != nil {
			return
		}
	}
	if t.verbose != 0 {
		fmt.Printf("%d iterations: %+v\n", iter, f)
	}
	for i := range objs {
		if objs[i] != 0 {
			f.Del(uint(i))
			objs[i] = 0
		}
		err = validate()
		if err != nil {
			return
		}
		iter++
	}
	if t.verbose != 0 {
		fmt.Printf("%d iterations: %+v\n", iter, f)
		fmt.Printf("No errors: %d iterations\n", t.iterations)
	}
	return
}
