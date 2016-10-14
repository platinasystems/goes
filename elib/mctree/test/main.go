package main

import (
	"github.com/platinasystems/go/elib/mctree"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ip4"

	"bufio"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

type ip4_route struct {
	Dst ip4.Prefix

	/* Next hop (host byte order). */
	Next_hop ip4.Address

	/* Route-views route type. */
	Type uint32
}

type test struct {
	route_views_path, load_path, save_path string

	cpu_profile_file               string
	load_tree_path, save_tree_path string

	// Instead of reading routes from route-views table generate sequential
	// table of the form 10.X/24 for incrementing X.
	n_sequential_slash_24_routes uint

	seed int64

	iter          uint
	n_iter        uint
	validate_iter uint
	print_iter    uint
	add_del_iter  uint
	verbose       uint

	mctree mctree.Main

	routes []ip4_route
}

func (t *test) parse_route_views(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	r_save := ip4_route{}
	n_subnets := 0
	routeIndex := 0
	for scanner.Scan() {
		line := scanner.Text()

		i := 0
		r := ip4_route{}
		r.Type = uint32(line[i])
		i++
		if line[i] == '*' {
			i++
		}

		var in parse.Input
		in.Add(string(line[i+1:]))

		var skip [2]int
		switch r.Type {
		case 'L', 'C':
			if !in.Parse("%v is directly connected,", &r.Dst) {
				panic(line)
			}
		case ' ':
			switch {
			case in.Parse("%v is variably subnetted, %d subnets", &r_save.Dst, &n_subnets):
			case in.Parse("%v is subnetted, %d subnets", &r_save.Dst, &n_subnets):
			default:
				panic(line)
			}
		case 'S', 'B':
			switch {
			case in.Parse("%v [%d/%d] via %v", &r.Dst, &skip[0], &skip[1], &r.Next_hop):
				t.routes = append(t.routes, r)
			case n_subnets > 0 && in.Parse("%v [%d/%d] via %v", &r.Dst.Address, &skip[0], &skip[1], &r.Next_hop):
				r.Dst.Len = r_save.Dst.Len
				n_subnets--
				t.routes = append(t.routes, r)
			default:
				panic(line)
			}
		default:
			panic(line)
		}
		routeIndex++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (t *test) save_routes(path string) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	e := gob.NewEncoder(f)
	e.Encode(t.routes)
}

func (t *test) load_routes(path string) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	d := gob.NewDecoder(f)
	d.Decode(&t.routes)
}

func (t *test) gen_sequential_slash_24_routes() {
	dst := &ip4.Address{10, 0, 0, 0}
	t.routes = make([]ip4_route, t.n_sequential_slash_24_routes)
	for i := uint(0); i < t.n_sequential_slash_24_routes; i++ {
		t.routes[i] = ip4_route{
			Dst: ip4.Prefix{Address: *dst, Len: 24},
		}
		dst[2]++
		if dst[2] == 0 {
			dst[1]++
			if dst[1] == 0 {
				dst[0]++
			}
		}
	}
}

func main() {
	t := &test{}

	cf := &t.mctree.Config
	flag.StringVar(&t.load_path, "load", "", "Path to routing table to load")
	flag.StringVar(&t.save_path, "save", "", "Path to routing table to save")
	flag.StringVar(&t.save_tree_path, "save-tree", "", "Path to save final tree to")
	flag.StringVar(&t.load_tree_path, "load-tree", "", "Path to load initial tree")
	flag.Int64Var(&t.seed, "seed", 0, "Seed for random number generator")
	flag.UintVar(&t.n_iter, "iter", 1, "Number of iterations to run")
	flag.UintVar(&cf.Validate_iter, "validate", 0, "Number of iterations between validate (0 means disable)")
	flag.UintVar(&t.print_iter, "print", 0, "Number of iterations between prints (0 means disable)")
	flag.UintVar(&t.add_del_iter, "add-del", 0, "Number of iterations between random add/del (0 means disable)")
	flag.UintVar(&t.verbose, "verbose", 0, "Verbosity level (0 means not verbose)")
	flag.UintVar(&cf.Restart_after_steps, "restart", 0, "Restart if no advance after this many steps")
	flag.UintVar(&cf.Max_leafs, "max-leafs", 16<<10, "Max number of leafs")
	flag.UintVar(&cf.Min_pairs_for_split, "min-pairs", 4, "Min number of pairs in a bucket for a split")
	flag.Float64Var(&cf.Temperature, "temperature", 1e-6, "Temperature for annealing")
	flag.StringVar(&t.cpu_profile_file, "cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&t.route_views_path, "route-views", "", `Input routing table from route-views.org "show ip route"`)
	flag.UintVar(&t.n_sequential_slash_24_routes, "24-routes", 0, "Use this many sequential 10.X/24 routes for routing table (instead of reading from file).")

	flag.Parse()

	if len(t.route_views_path)+len(t.load_path)+int(t.n_sequential_slash_24_routes) == 0 {
		log.Fatalf("no input file")
	}

	switch {
	case len(t.route_views_path) > 0:
		t.parse_route_views(t.route_views_path)
	case t.n_sequential_slash_24_routes > 0:
		t.gen_sequential_slash_24_routes()
	case len(t.load_path) > 0:
		t.load_routes(t.load_path)
	}

	if len(t.save_path) > 0 {
		t.save_routes(t.save_path)
		return
	}

	if t.seed == 0 {
		t.seed = int64(os.Getpid())
	}
	rand.Seed(t.seed)

	if true {
		if t.cpu_profile_file != "" {
			f, err := os.Create(t.cpu_profile_file)
			if err != nil {
				log.Fatal(err)
			}
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}

		t.optimize_table()
	}
}

func ip4_pair(x *ip4.Prefix, p *mctree.Pair) {
	v, m := vnet.Uint32(x.Address.AsUint32()), x.Mask()
	p.Set(uint(v), uint(m))
	if p.Value&^p.Mask != 0 {
		panic(fmt.Errorf("bad pair %s", p))
	}
}

func (test *test) optimize_table() {
	m := &test.mctree

	m.Config.Key_bits = ip4.AddressBits

	m.Init()
	if len(test.load_tree_path) > 0 {
		if err := m.Restore(test.load_tree_path); err != nil {
			panic(err)
		}
	}
	for i := range test.routes {
		r := &test.routes[i]
		var p [1]mctree.Pair
		if r.Dst.Len == 0 { // ignore default route
			continue
		}
		ip4_pair(&r.Dst, &p[0])
		m.AddDel(p[:], false)
	}

	fmt.Println("seed: ", test.seed)

	defer func() {
		if e := recover(); e != nil {
			panic(fmt.Errorf("validate fails iter %d: %s", test.iter, e))
		}
	}()

	if m.Validate_iter != 0 {
		m.Validate()
	}
	start := time.Now()
	m.Print(0, start, test.verbose != 0)

	for test.iter = 0; test.iter < test.n_iter; test.iter++ {
		i := test.iter
		if test.print_iter != 0 && i%test.print_iter == 0 {
			m.Print(i, start, test.verbose != 0)
		}

		if lower_cost_found := m.Step(); lower_cost_found {
		}

		if m.Validate_iter != 0 && i%m.Validate_iter == 0 {
			m.Validate()
		}

		if test.add_del_iter != 0 && i%test.add_del_iter == 0 {
			r := &test.routes[rand.Intn(len(test.routes))]
			var p [1]mctree.Pair
			ip4_pair(&r.Dst, &p[0])
			b := uint32(1) << 31
			is_del := r.Type&b == 0
			m.AddDel(p[:], is_del)
			r.Type ^= b

			if m.Validate_iter != 0 && i%m.Validate_iter == 0 {
				m.Validate()
			}
		}
	}

	m.Print(test.iter, start, test.verbose != 0)

	if len(test.save_tree_path) > 0 {
		if err := m.Save(test.save_tree_path); err != nil {
			panic(err)
		}
		return
	}
}
