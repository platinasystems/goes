// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elib

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

func (s *Sparse) validate() (err error) {
	if len(s.count) != len(s.valid) {
		err = fmt.Errorf("len mismatch %d != %d", len(s.count), len(s.valid))
		return
	}

	nValid := int32(0)
	for i := range s.count {
		if nValid != s.count[i] {
			err = fmt.Errorf("n valid %d != count[%d] %d", nValid, i, s.count[i])
			return
		}
		nValid += int32(NSetBits(Word(s.valid[i])))
	}

	return
}

type testSparse struct {
	// Number of iterations to run
	iterations Count

	// Validate/print every so many iterations (zero means never).
	validateEvery Count
	printEvery    Count

	// Seed to make randomness deterministic.  0 means choose seed.
	seed int64

	// Number of objects to create.
	nObjects Count

	// Number of bits in sparse index.
	log2SparseIndexMax uint

	verbose int

	x       Sparse
	objs    randSparseVec
	toDense map[Index]Index
	unique  []Word

	profile string
}

type randSparse struct {
	sparse, dense Index
}

//go:generate gentemplate -d Package=elib -id randSparse -d VecType=randSparseVec -d Type=randSparse vec.tmpl

func SparseTest() {
	t := testSparse{
		iterations:    10,
		nObjects:      10,
		validateEvery: 1,
		verbose:       1,
	}
	flag.Var(&t.iterations, "iter", "Number of iterations")
	flag.Var(&t.validateEvery, "valid", "Number of iterations per validate")
	flag.Var(&t.printEvery, "print", "Number of iterations per print")
	flag.Int64Var(&t.seed, "seed", 0, "Seed for random number generator")
	flag.UintVar(&t.log2SparseIndexMax, "len", 16, "Number of bits in sparse index")
	flag.Var(&t.nObjects, "objects", "Number of random objects")
	flag.IntVar(&t.verbose, "verbose", 0, "Be verbose")
	flag.StringVar(&t.profile, "profile", "", "Write CPU profile to file")
	flag.Parse()

	err := runSparseTest(&t)
	if err != nil {
		panic(err)
	}
}

func (t *testSparse) validate(iter int) (err error) {
	if t.validateEvery != 0 && iter%int(t.validateEvery) == 0 {
		if err = t.x.validate(); err != nil {
			if t.verbose != 0 {
				fmt.Printf("iter %d: %s\n%+v\n", iter, err, t.x)
			}
			return
		}
	}

	for i := 0; i < len(t.objs); i++ {
		o := &t.objs[i]
		d, ok := t.toDense[o.sparse]
		if !ok {
			panic("toDense corrupt")
		}
		if d != o.dense {
			err = fmt.Errorf("sparse 0x%x dense %d != %d", o.sparse, o.dense, d)
			return
		}
	}

	for i := 0; i < len(t.objs); i++ {
		o := &t.objs[i]
		d, ok := t.x.Get(o.sparse)
		if ok != (o.dense != MaxIndex) {
			err = fmt.Errorf("ok %v sparse 0x%x dense %d != %d",
				iter, ok, o.sparse, o.dense, d)
			return
		}
	}
	return
}

func (t *testSparse) randomSparse() (i Index) {
	for {
		i = Index(rand.Int() & (1<<uint(t.log2SparseIndexMax) - 1))
		if !bitmapSet(t.unique, uint(i)) {
			return
		}
	}
}

func runSparseTest(t *testSparse) (err error) {
	if t.seed == 0 {
		t.seed = int64(time.Now().Nanosecond())
	}

	rand.Seed(t.seed)
	if t.verbose != 0 {
		fmt.Printf("%#v\n", t)
	}
	t.objs.Resize(uint(t.nObjects))
	t.unique = bitmapMake(1 << t.log2SparseIndexMax)

	t.toDense = make(map[Index]Index)

	for i := range t.objs {
		t.objs[i].sparse = t.randomSparse()
		t.objs[i].dense = MaxIndex
		t.toDense[t.objs[i].sparse] = t.objs[i].dense
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

	var iter int
	for iter = 0; iter < int(t.iterations); iter++ {
		oi := rand.Int() % len(t.objs)
		o := &t.objs[oi]
		if o.dense != MaxIndex {
			if ok := t.x.Unset(o.sparse); !ok {
				err = fmt.Errorf("not found 0x%x", o.sparse)
				return
			}
			o.dense = MaxIndex
		} else {
			bitmapUnset(t.unique, uint(o.sparse))
			if t.validateEvery != 0 {
				delete(t.toDense, o.sparse)
			}
			o.sparse = t.randomSparse()
			o.dense = t.x.Set(t.objs[oi].sparse)
		}
		t.objs[oi] = *o
		if t.validateEvery != 0 {
			t.toDense[o.sparse] = o.dense
		}
		if t.validateEvery != 0 && iter%int(t.validateEvery) == 0 {
			err = t.validate(iter)
			if err != nil {
				err = fmt.Errorf("iter %d %v", iter, err)
				return
			}
		}
		if t.printEvery != 0 && iter%int(t.printEvery) == 0 {
			fmt.Printf("%10g iterations: %s\n", float64(iter), &t.x)
		}
	}
	if t.verbose != 0 {
		fmt.Printf("%d iterations: %+v\n", iter, t.x)
	}

	for i := range t.objs {
		o := &t.objs[i]
		if o.dense == MaxIndex {
			continue
		}
		if ok := t.x.Unset(o.sparse); !ok {
			err = fmt.Errorf("not found 0x%x", o.sparse)
			return
		}
		o.dense = MaxIndex
		if t.validateEvery != 0 {
			t.toDense[o.sparse] = o.dense
			if iter%int(t.validateEvery) == 0 {
				err = t.validate(iter)
				if err != nil {
					return
				}
			}
		}
		if t.printEvery != 0 && iter%int(t.printEvery) == 0 {
			fmt.Printf("%10g iterations: %s\n", float64(iter), &t.x)
		}
		iter++
	}
	if t.verbose != 0 {
		fmt.Printf("%d iterations: %+v\n", iter, t.x)
		fmt.Printf("No errors: %d iterations\n", t.iterations)
	}
	return
}
