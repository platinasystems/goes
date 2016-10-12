package mctree

import (
	"fmt"
	"sort"
)

type maxPair [8]Pair

func newMaxPair(p []Pair) (m maxPair) {
	if copy(m[:], p) < len(p) {
		panic("must increase max_words_per_key")
	}
	return
}

func (p *maxPair) is_equal(q []Pair) bool {
	for i := range q {
		if !p[i].equal(&q[i]) {
			return false
		}
	}
	return true
}

type validate_tree struct {
	validate_pair_counts map[maxPair]int
}

type validate_main struct {
	validate_active_nodes     map[node_index]bool
	validate_referenced_pairs map[uint]uint
	validate_all_pairs        map[maxPair]bool
}

func (c *Config) wantValidate() bool { return c.Validate_iter > 0 }

func (m *validate_main) add_del_pair(p []Pair, is_del bool) {
	if m.validate_all_pairs == nil {
		return
	}
	mp := newMaxPair(p)
	if is_del {
		delete(m.validate_all_pairs, mp)
	} else {
		m.validate_all_pairs[mp] = true
	}
}

func (t *tree) validate_node(i node_index, key *maxPair, cost *tree_cost, n_pairs_per_key uint) {
	n := t.get_node(i)
	nkey := t.get_node_key(i)
	if !key.is_equal(nkey) {
		panic(fmt.Errorf("key %x != %x", nkey, key[:len(nkey)]))
	}
	t.validate_active_nodes[i] = true
	if n.is_split() {
		if n.n_pairs() != 0 {
			panic(fmt.Errorf("split node must have no pairs"))
		}

		var sk [2]maxPair
		b0, b1 := index(n.split_bit)
		sk[0] = *key
		sk[0][b0].Mask |= b1
		sk[1] = sk[0]
		sk[1][b0].Value |= b1

		for j := range n.sub_nodes {
			t.validate_node(n.sub_nodes[j], &sk[j], cost, n_pairs_per_key)
		}
	} else {
		po := &n.pair_offsets
		if po.pool_index != invalid {
			t.validate_referenced_pairs[po.pool_index] += 1
		}
		for i, o := range po.vec {
			if n.index_is_free(uint(i)) {
				continue
			}
			p := t.pair_hash.get_pairs_for_offset(o)
			var px maxPair
			for j := range p {
				if p[j].Value&p[j].Mask&nkey[j].Mask != nkey[j].Value&p[j].Mask {
					panic(fmt.Errorf("wrong sub node"))
				}
				px[j] = p[j]
			}
			t.validate_pair_counts[px] += 1
		}
		x := float64(n.n_pairs())
		cost.occupancy += x
		cost.occupancy2 += x * x
	}
}

func (m *Main) validate_tree(t *tree) {
	if t.validate_pair_counts == nil {
		t.validate_pair_counts = make(map[maxPair]int)
	}
	t.Main = m

	var c tree_cost
	t.validate_node(t.root_node_index, &maxPair{}, &c, m.n_pairs_per_key)

	// Validate cost
	if t.occupancy != c.occupancy {
		panic(fmt.Errorf("occupancy mismatch got %.0f != want %.0f", t.occupancy, c.occupancy))
	}
	if t.occupancy2 != c.occupancy2 {
		panic(fmt.Errorf("occupancy2 mismatch got %.0f != want %.0f", t.occupancy2, c.occupancy2))
	}
	c.compute_cost(m)
	if t.cost != c.cost {
		panic(fmt.Errorf("cost mismatch got %.0f != want %.0f", t.cost, c.cost))
	}

	if got, want := len(t.validate_pair_counts), len(m.validate_all_pairs); got != want {
		panic(fmt.Errorf("pair leak got %d != want %d", got, want))
	}
	for k := range m.validate_all_pairs {
		if _, ok := t.validate_pair_counts[k]; !ok {
			panic(fmt.Errorf("pair %s vanished", &k))
		}
		delete(t.validate_pair_counts, k)
	}
}

func (m *Main) Validate() {
	if m.validate_active_nodes == nil {
		m.validate_active_nodes = make(map[node_index]bool)
		m.validate_referenced_pairs = make(map[uint]uint)
	}
	t := m.get_tree_seq(m.tree_sequence)
	if t.n_steps > 0 {
		m.validate_tree(t)
	}
	m.validate_tree(m.get_tree_seq(m.tree_sequence - 1))
	if want, got := m.node_pool.Elts(), uint(len(m.validate_active_nodes)); want != got {
		panic(fmt.Errorf("wrong number of active nodes want %d != got %d", want, got))
	}
	for ni := range m.validate_active_nodes {
		if m.node_pool.IsFree(uint(ni)) {
			panic("node pool pointer to free node")
		}
		delete(m.validate_active_nodes, ni)
	}

	if got, want := uint(len(m.validate_referenced_pairs)), m.shared_pair_offsets_pool.Elts(); got != want {
		panic(fmt.Errorf("node referenced pairs %d != pair vec pool elts %d", got, want))
	}
	for i := range m.nodes {
		if m.node_pool.IsFree(uint(i)) {
			continue
		}
		n := m.get_node(node_index(i))
		po := &n.pair_offsets
		if po.pool_index == invalid {
			continue
		}
		if got, want := m.shared_pair_offsets_pool.elts[po.pool_index].reference_count, m.validate_referenced_pairs[po.pool_index]; got != want {
			panic(fmt.Errorf("ref count mismatch %d != %d", got, want))
		}
	}
	for i := range m.validate_referenced_pairs {
		delete(m.validate_referenced_pairs, i)
	}
}

type dump_node struct {
	key     maxPair
	n_pairs uint32
	level   uint32
}

//go:generate gentemplate -d Package=mctree -id dump_node -d VecType=dump_node_vec -d Type=dump_node github.com/platinasystems/go/elib/vec.tmpl

func (d *dump_node_vec) dump_node(m *Main, i node_index, level uint32) {
	n := m.get_node(i)
	if n.is_split() {
		d.dump_node(m, n.sub_nodes[0], level+1)
		d.dump_node(m, n.sub_nodes[1], level+1)
	} else {
		*d = append(*d, dump_node{
			level:   level,
			key:     newMaxPair(m.get_node_key(i)),
			n_pairs: uint32(n.n_pairs()),
		})
	}
}

func (d *dump_node_vec) dump(m *Main, t *tree) {
	d.dump_node(m, t.root_node_index, 0)
}

type dump_node_sort dump_node_vec

func (d dump_node_sort) Len() int { return len(d) }
func (d dump_node_sort) Less(i, j int) bool {
	for k := range d[i].key {
		if d[i].key[k].Value < d[j].key[k].Value {
			return true
		}
	}
	return false
}
func (d dump_node_sort) Swap(i, j int) { d[i], d[j] = d[j], d[i] }

func (t *tree) dump(m *Main) {
	d := dump_node_vec{}
	d.dump(m, t)
	sort.Sort(dump_node_sort(d))
	for _, x := range d {
		fmt.Printf("[%d] %s %d leafs\n", x.level, &x.key, x.n_pairs)
	}
}
