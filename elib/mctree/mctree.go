package mctree

import (
	"github.com/platinasystems/go/elib"

	"fmt"
	"math"
	"math/rand"
	"time"
)

type shared_pair_offsets struct {
	vec             pair_offset_vec
	reference_count uint
}

//go:generate gentemplate -d Package=mctree -id shared_pair_offsets_pool -d PoolType=shared_pair_offsets_pool -d Type=shared_pair_offsets -d Data=elts github.com/platinasystems/go/elib/pool.tmpl

type pair_offsets struct {
	m          *Main
	vec        pair_offset_vec
	pool_index uint
	hash       elib.Hash
	hash_valid bool
}

func (p *pair_offsets) Len() uint { return p.vec.Len() }

func (o pair_offset) HashKey(s *elib.HashState) { s.HashUint64(uint64(o), 0, 0, 0) }
func (a pair_offset) HashKeyEqual(h elib.Hasher, i uint) bool {
	po := h.(*pair_offsets)
	return a == pair_offset(po.vec[i])
}
func (p *pair_offsets) HashIndex(s *elib.HashState, i uint) { s.HashUint64(uint64(p.vec[i]), 0, 0, 0) }
func (p *pair_offsets) HashResize(newCap uint, rs []elib.HashResizeCopy) {
	var d pair_offsets
	m := p.m
	d.get(m, newCap)
	for i := range d.vec {
		d.vec[i] = pair_offset_invalid
	}
	for i := range rs {
		d.vec[rs[i].Dst] = p.vec[rs[i].Src]
	}
	if p.pool_index != invalid {
		p.put(m)
	}
	// Replace with new vector/pool index.
	p.vec = d.vec
	p.pool_index = d.pool_index
}

const invalid = ^uint(0)

func (p *pair_offsets) get(m *Main, l uint) {
	i := invalid
	var v pair_offset_vec
	if l > 0 {
		i = m.shared_pair_offsets_pool.GetIndex()
		sp := &m.shared_pair_offsets_pool.elts[i]
		sp.vec.Validate(uint(l - 1))
		v = sp.vec[:l]
		sp.reference_count = 1
	}
	p.vec = v
	p.m = m
	p.pool_index = i
}

func (p *pair_offsets) put(m *Main) (freed bool) {
	freed = m.put_pair_offsets(p.vec, p.pool_index)
	return
}

func (m *Main) put_pair_offsets(vec pair_offset_vec, pool_index uint) (freed bool) {
	if m.shared_pair_offsets_pool.IsFree(pool_index) {
		panic("put free index")
	}
	sp := &m.shared_pair_offsets_pool.elts[pool_index]
	if sp.reference_count == 0 {
		panic("ref count 0")
	}
	sp.reference_count--
	if sp.reference_count == 0 {
		freed = true
		sp.vec = vec
		m.shared_pair_offsets_pool.PutIndex(pool_index)
	}
	return
}

func (p *pair_offsets) add_reference(m *Main) {
	if p.pool_index == invalid {
		return
	}
	if m.shared_pair_offsets_pool.IsFree(p.pool_index) {
		panic("ref free index")
	}
	sp := &m.shared_pair_offsets_pool.elts[p.pool_index]
	sp.reference_count++
}

func (p *pair_offsets) invalidate() {
	p.hash_valid = false
	p.vec = nil
	p.pool_index = invalid
}

type node_index uint32

const invalid_node_index = ^node_index(0)

// Fields to copy when cloning a node.
type node_clone_copy struct {
	// Common mask and value for all sub-pairs (pairs under this node).
	// All sub-pairs P will satisfy KEY.mask & P.value == KEY.value
	key_offset pair_offset

	// All pairs with split bit unmasked or set to 0 are copied to sub node 0;
	// all pairs with split bit unmasked or set to 1 are copied to sub node 1.
	split_bit uint
}

func (m *Main) get_node_key(i node_index) []Pair {
	return m.node_keys.get_pairs_for_index(uint(i), m.n_pairs_per_key)
}
func (m *Main) validate_node_key(i node_index) {
	m.node_keys.Validate(uint(i+1)*m.n_pairs_per_key - 1)
}

type node struct {
	node_clone_copy

	// Index within node pool.
	index node_index

	// Sub-nodes with split bit matched as 0 and/or 1.
	sub_nodes [2]node_index

	// Offsets of pairs attached to this node.  Only non-empty when this node has no sub nodes.
	pair_offsets
}

func (p *pair_offsets) hash_invalidate() {
	if !p.hash_valid {
		return
	}
	p.hash_valid = false
	n := 0
	for i := range p.vec {
		if !p.hash.IsFree(uint(i)) {
			p.vec[n] = p.vec[i]
			n++
		}
	}
	p.hash.Clear()
}

func (p *pair_offsets) hash_validate(m *Main) {
	h := &p.hash

	// Already valid?
	if p.hash_valid {
		h.Hasher = p
		return
	}
	p.hash_valid = true

	// Add a reference so offset vector will not be reused by hash resize.
	save_vec, save_pool_index := p.vec, p.pool_index
	p.add_reference(m)

	if h.Hasher == nil {
		h.Init(p, p.vec.Len())
	} else {
		h.Hasher = p
		h.Clear()
	}
	for _, o := range save_vec {
		i, _ := h.Set(o)
		p.vec[i] = o
	}
	if save_pool_index != invalid {
		m.put_pair_offsets(save_vec, save_pool_index)
	}
}

//go:generate gentemplate -d Package=mctree -id node -d PoolType=node_pool -d Type=node -d Data=nodes github.com/platinasystems/go/elib/pool.tmpl

func (m *Main) get_node(i node_index) *node {
	return &m.node_pool.nodes[i]
}

func (m *Main) new_node(seq uint32) (ni node_index) {
	ni = node_index(m.node_pool.GetIndex())
	n := m.get_node(ni)
	n.index = ni
	n.sub_nodes[0] = invalid_node_index
	n.sub_nodes[1] = invalid_node_index
	n.pair_offsets.invalidate()
	m.validate_node_key(ni)
	return
}

func (m *Main) free_node(n *node) {
	n.free_pairs(m)
	m.node_pool.PutIndex(uint(n.index))
	n.sub_nodes[0] = invalid_node_index
	n.sub_nodes[1] = invalid_node_index
}

func (n *node) get_subs(m *Main) (n0, n1 *node) {
	n0, n1 = m.get_node(n.sub_nodes[0]), m.get_node(n.sub_nodes[1])
	return
}

func index(bit uint) (b0, b1 word) {
	b0, b1 = word(bit/word_bits), word(1)<<(bit%word_bits)
	return
}

func (n *node) alloc_subs(m *Main, bit uint) (new_n, n0, n1 *node) {
	ni0 := m.new_node(m.tree_sequence)
	ni1 := m.new_node(m.tree_sequence)
	n0, n1 = m.get_node(ni0), m.get_node(ni1)
	n = m.get_node(n.index) // re-fetch since address can change
	new_n = n

	n.sub_nodes[0] = ni0
	n.sub_nodes[1] = ni1
	n.split_bit = bit

	k, k0, k1 := m.get_node_key(n.index), m.get_node_key(ni0), m.get_node_key(ni1)
	b0, b1 := index(bit)
	copy(k0, k)
	copy(k1, k)
	k0[b0].Mask |= b1
	k1[b0].Mask |= b1
	k1[b0].Value |= b1
	return
}

func (n *node) free_pairs(m *Main) {
	if n.pair_offsets.pool_index != invalid {
		if n.pair_offsets.put(m) {
			n.pair_offsets.hash_invalidate()
		}
	}
	n.pair_offsets.invalidate()
}

func (n *node) is_split() bool {
	return n.sub_nodes[0] != invalid_node_index && n.sub_nodes[1] != invalid_node_index
}

func (n *node) n_pairs() uint {
	p := &n.pair_offsets
	if p.hash_valid {
		return p.hash.Elts()
	} else {
		return p.vec.Len()
	}
}

func (n *node) index_is_free(i uint) bool {
	p := &n.pair_offsets
	return p.hash_valid && p.hash.IsFree(uint(i))
}

type tree_cost struct {
	// Total number of pairs in all leafs.
	// A pair may be in multiple leafs so this is always >= number of pairs.
	occupancy float64

	// Sum of squares of leaf occupancy.  Used to compute cost.
	occupancy2 float64

	// Number of non-empty leafs.
	n_non_empty_leafs float64

	cost float64
}

// Measures log2 (current/ideal) occupancy.
func (c *tree_cost) compute_q(m *Main) (q float64) {
	have := c.occupancy / c.n_non_empty_leafs
	np := m.n_pairs()
	max_leafs := m.Max_leafs
	if max_leafs > np {
		max_leafs = np
	}
	ideal := float64(np) / float64(max_leafs)
	q = math.Log2(have / ideal)
	return
}

func (c *tree_cost) compute_cost(m *Main) {
	n := m.n_pairs()
	if n == 0 {
		c.cost = 0
	} else {
		np := float64(n)
		c.cost = c.occupancy2 / (np * np)
		c.cost *= c.occupancy / np
		if c.cost < 0 {
			panic("negative cost")
		}
	}
}

func (c *tree_cost) add_del_occupancy(m *Main, l uint, isDel bool) {
	if isDel {
		// Occupancy^2 decreases from l^2 to (l-1)^2 = l^2 - 2l + 1
		c.occupancy -= 1
		c.occupancy2 += 1 - 2*float64(l)
		if l == 1 {
			c.n_non_empty_leafs -= 1
		}
	} else {
		if l == 0 {
			c.n_non_empty_leafs += 1
		}
		// Occupancy^2 increases from l^2 to (l+1)^2 = l^2 + 2l + 1
		c.occupancy += 1
		c.occupancy2 += 1 + 2*float64(l)
	}
	c.compute_cost(m)
}

func (m *Main) accept_cost_change(dcost float64) (accept bool) {
	rnd := rand.Float64()
	exp := math.Exp(-dcost / m.temperature)
	accept = rnd < exp
	return
}

type tree struct {
	*Main

	root_node_index node_index

	n_steps uint

	tree_cost

	validate_tree
}

type step_stats struct {
	attempted uint64
	accepted  uint64
	advanced  uint64
}

type Config struct {
	Key_bits            uint
	Restart_after_steps uint
	Max_leafs           uint
	Min_pairs_for_split uint
	Validate_iter       uint
	Temperature         float64
}

type Main struct {
	// Allocation pool of nodes.
	node_pool

	node_keys pair_vec

	// Size of key in units of Word (32 bits).
	// For ip4 set to 1.  For ip4/64 set to 2.
	n_pairs_per_key uint

	pair_hash pair_hash

	shared_pair_offsets_pool

	// Temperature for simulated annealing.
	temperature float64

	// Incremented for each tree accepted with a lower cost than before.
	// Current tree we are working on is always trees[tree_sequence&1].
	// Previous tree is always trees[(tree_sequence-1)&1].
	tree_sequence uint32

	// Current trial tree & previous (lowest cost) tree.
	// Indexed by low bit of tree sequence number.
	trees [2]tree

	random_bit_buffer

	stats struct {
		split, join step_stats
		n_restart   uint64
	}

	Config

	validate_main
}

func (m *Main) n_pairs() uint                 { return m.pair_hash.hash.Elts() }
func (m *Main) get_tree_seq(seq uint32) *tree { return &m.trees[seq&1] }
func (m *Main) get_tree() *tree               { return m.get_tree_seq(m.tree_sequence) }
func (m *Main) get_min_tree() *tree {
	seq := m.tree_sequence
	if m.tree_sequence > 0 {
		seq--
	}
	return m.get_tree_seq(seq)
}

func branch(p []Pair, b0, b1 uint) (r0, r1 word) {
	if (p[b0].Mask>>b1)&1 == 0 {
		r0, r1 = 1, 1
	} else {
		r1 = (p[b0].Value >> b1) & 1
		r0 = ^r1 & 1
	}
	return
}

func (sup *node) split(m *Main, bit uint) (accept_split bool) {
	m.stats.split.attempted++
	if sup.is_split() {
		panic("already split")
	}
	var (
		ps [2]pair_offsets
		l  uint
	)
	if l = sup.n_pairs(); l == 0 {
		return
	}

	ps[0].get(m, l)
	ps[1].get(m, l)

	b0, b1 := bit/word_bits, bit%word_bits
	i0, i1 := word(0), word(0)
	for i, o := range sup.pair_offsets.vec {
		if sup.index_is_free(uint(i)) {
			continue
		}
		p := m.pair_hash.get_pairs_for_offset(o)
		r0, r1 := branch(p, b0, b1)
		ps[0].vec[i0] = o
		ps[1].vec[i1] = o
		i0 += r0
		i1 += r1
	}
	ps[0].vec = ps[0].vec[:i0]
	ps[1].vec = ps[1].vec[:i1]

	fl, fi0, fi1 := float64(l), float64(i0), float64(i1)

	t := m.get_tree()
	c := t.tree_cost

	c.n_non_empty_leafs -= 1
	if i0 > 0 {
		c.n_non_empty_leafs += 1
	}
	if i1 > 0 {
		c.n_non_empty_leafs += 1
	}
	c.occupancy += fi0 + fi1 - fl
	c.occupancy2 += fi0*fi0 + fi1*fi1 - fl*fl
	c.compute_cost(m)

	dcost := c.cost - t.cost
	accept_split = dcost < 0
	if accept_split {
		m.stats.split.advanced++
	} else if i0 > 0 && i1 > 0 {
		accept_split = m.accept_cost_change(dcost)
	}
	if !accept_split {
		ps[0].put(m)
		ps[1].put(m)
		return
	}

	m.stats.split.accepted++
	sup, n0, n1 := sup.alloc_subs(m, bit)
	n0.pair_offsets = ps[0]
	n1.pair_offsets = ps[1]
	n0.pair_offsets.hash_invalidate()
	n1.pair_offsets.hash_invalidate()
	sup.free_pairs(m)
	t.tree_cost = c
	return
}

func (n *node) join(m *Main) (accept_join bool) {
	m.stats.join.attempted++

	if !n.is_split() {
		return
	}
	n0, n1 := n.get_subs(m)
	if n0.is_split() || n1.is_split() {
		panic("join split node")
	}

	var pos pair_offsets
	l0, l1 := n0.n_pairs(), n1.n_pairs()
	pos.get(m, uint(l0+l1))

	// Copy first node offsets into place.
	npos := uint(0)
	po0, po1 := &n0.pair_offsets, &n1.pair_offsets
	if po0.hash_valid {
		for i, o := range n0.pair_offsets.vec {
			if !n0.pair_offsets.hash.IsFree(uint(i)) {
				pos.vec[npos] = o
				npos++
			}
		}
	} else {
		copy(pos.vec[:l0], n0.pair_offsets.vec[:l0])
		npos = l0
	}

	// Copy second node offsets into place (but only if masked).
	if po1.hash_valid {
		for i, o := range n1.pair_offsets.vec {
			if !po1.hash.IsFree(uint(i)) &&
				m.pair_hash.is_masked(o, n.split_bit) {
				pos.vec[npos] = o
				npos++
			}
		}
	} else {
		for _, o := range n1.pair_offsets.vec {
			if m.pair_hash.is_masked(o, n.split_bit) {
				pos.vec[npos] = o
				npos++
			}
		}
	}
	l := npos
	pos.vec = pos.vec[:l]

	fl, fl0, fl1 := float64(l), float64(l0), float64(l1)

	t := m.get_tree()
	c := t.tree_cost

	if l > 0 {
		c.n_non_empty_leafs += 1
	}
	if l0 > 0 {
		c.n_non_empty_leafs -= 1
	}
	if l1 > 0 {
		c.n_non_empty_leafs -= 1
	}
	c.occupancy += fl - fl0 - fl1
	c.occupancy2 += fl*fl - fl0*fl0 - fl1*fl1
	c.compute_cost(m)

	dcost := c.cost - t.cost
	accept_join = dcost < 0
	if accept_join {
		m.stats.join.advanced++
	} else {
		accept_join = m.accept_cost_change(dcost)
	}

	if !accept_join {
		pos.put(m)
		return
	}

	m.stats.join.accepted++
	n.pair_offsets = pos
	n.pair_offsets.hash_invalidate()
	t.tree_cost = c

	n.split_bit = 0
	n.sub_nodes[0] = invalid_node_index
	n.sub_nodes[1] = invalid_node_index
	m.free_node(n0)
	m.free_node(n1)
	return
}

func (m *Main) free_tree_node(i node_index) {
	n := m.get_node(i)
	if n.is_split() {
		m.free_tree_node(n.sub_nodes[0])
		m.free_tree_node(n.sub_nodes[1])
	}
	m.free_node(n)
}

func (m *Main) free_tree(seq uint32) {
	t := m.get_tree_seq(seq)
	m.free_tree_node(t.root_node_index)
}

func (m *Main) clone_tree_node(dst_seq uint32, di, si node_index) {
	if m.nodes[si].is_split() {
		for i := range m.nodes[di].sub_nodes {
			ni := m.new_node(dst_seq)
			m.nodes[di].sub_nodes[i] = ni
			m.clone_tree_node(dst_seq, ni, m.nodes[si].sub_nodes[i])
		}
	}

	m.nodes[si].pair_offsets.add_reference(m)
	m.nodes[di].pair_offsets = m.nodes[si].pair_offsets
	m.nodes[di].node_clone_copy = m.nodes[si].node_clone_copy

	{
		kd, ks := m.get_node_key(di), m.get_node_key(si)
		copy(kd, ks)
	}
}

func (m *Main) clone_tree(dst_seq, src_seq uint32) {
	dst := m.get_tree_seq(dst_seq)
	src := m.get_tree_seq(src_seq)
	dst.root_node_index = m.new_node(dst_seq)
	m.clone_tree_node(dst_seq, dst.root_node_index, src.root_node_index)
	dst.tree_cost = src.tree_cost
	dst.n_steps = 0
}

func (n *node) random_masked_bit(m *Main) (bit uint, ok bool) {
	po := &n.pair_offsets
	var o pair_offset
	if po.hash_valid {
		ri := po.hash.RandIndex()
		o = po.vec[ri]
	} else {
		o = po.vec[rand.Intn(int(po.Len()))]
	}

	p := m.pair_hash.get_pairs_for_offset(o)
	k := m.get_node_key(n.index)

	// Check that there are masked bits not already masked in node's key.
	for i := range p {
		if ok = p[i].Mask&^k[i].Mask != 0; ok {
			break
		}
	}
	if !ok {
		return
	}

	for {
		// Choose random word with bits that are masked but not in node's key.
		i := uint(rand.Intn(int(m.n_pairs_per_key)))
		mi := elib.Word(p[i].Mask &^ k[i].Mask)
		if mi == 0 {
			continue
		}

		// Chose random set bit.
		bi := rand.Intn(int(mi.NSetBits()))
		for {
			f := mi.FirstSet()
			if bi == 0 {
				bit = i*word_bits + f.MinLog2()
				return
			}
			bi--
			mi ^= f
		}
	}
	return
}

func (m *Main) random_leaf() (n, parent *node) {
	t := m.get_tree()
	i := t.root_node_index
	for {
		n = m.get_node(i)
		if !n.is_split() {
			break
		}
		parent = n
		i = parent.sub_nodes[m.random_bit()]
	}
	return
}

func (m *Main) is_joinable(sub, sup *node) (ok bool) {
	if sup != nil {
		n0, n1 := sup.get_subs(m)
		ok = !n0.is_split() && !n1.is_split()
	}
	return
}

func (m *Main) new_root(l uint) (n *node, t *tree) {
	t = m.get_tree()
	t.root_node_index = m.new_node(m.tree_sequence)
	n = m.get_node(t.root_node_index)
	n.pair_offsets.get(m, l)
	if l > 0 {
		t.n_non_empty_leafs = 1
		t.occupancy = float64(l)
		t.occupancy2 = t.occupancy * t.occupancy
		t.compute_cost(m)
	}
	return
}

func (m *Main) add_del_key_leaf(t *tree, key []Pair, node *node, is_del bool) {
	po := &node.pair_offsets

	po.hash_validate(m)

	l := po.hash.Elts()

	var (
		i      uint
		o      pair_offset
		exists bool
	)
	if o, exists = m.pair_hash.get(key); !exists {
		panic("key not found")
	}
	if is_del {
		i, exists = po.hash.Unset(o)
		if !exists {
			panic("offset not found")
		}
		po.vec[i] = pair_offset_invalid
	} else {
		if po.pool_index == invalid {
			po.get(m, po.hash.Cap())
		}
		i, _ := po.hash.Set(o)
		po.vec.Validate(i)
		po.vec[i] = o
	}
	t.add_del_occupancy(m, l, is_del)
}

func (m *Main) add_del_key_helper(t *tree, key []Pair, ni node_index, is_del bool) {
	node := m.get_node(ni)
	if node.is_split() {
		b0, b1 := node.split_bit/word_bits, node.split_bit%word_bits
		unmasked := (key[b0].Mask>>b1)&1 == 0
		v := (key[b0].Value>>b1)&1 != 0
		if unmasked || !v {
			m.add_del_key_helper(t, key, node.sub_nodes[0], is_del)
		}
		if unmasked || v {
			m.add_del_key_helper(t, key, node.sub_nodes[1], is_del)
		}
	} else {
		m.add_del_key_leaf(t, key, node, is_del)
	}
}

func (m *Main) AddDel(p []Pair, is_del bool) {
	// Cancels any optimizing steps currently in progress.
	m.restart()
	t := m.get_tree_seq(m.tree_sequence - 1)

	if !is_del {
		if _, exists := m.pair_hash.set(p); !exists {
			if m.validate_all_pairs != nil {
				m.validate_all_pairs[newMaxPair(p)] = true
			}
		}
	}

	m.add_del_key_helper(t, p, t.root_node_index, is_del)

	if is_del {
		m.pair_hash.unset(p)
		if m.validate_all_pairs != nil {
			delete(m.validate_all_pairs, newMaxPair(p))
		}
	}

	t.compute_cost(m)
	m.validate_main.add_del_pair(p, is_del)
}

func (m *Main) restart() {
	t := m.get_tree()
	if t.n_steps > 0 {
		m.free_tree(m.tree_sequence)
		t.n_steps = 0
	}
}

func (m *Main) Step() (lower_cost_found bool) {
	t := m.get_tree()

	// Clone tree on first step; each step edit tree trying to find a lower cost.
	if t.n_steps == 0 {
		m.clone_tree(m.tree_sequence, m.tree_sequence-1)
	}

	max_leafs := m.Max_leafs
	// Never allow more leafs than we have pairs.
	if np := m.n_pairs(); np < max_leafs {
		max_leafs = np
	}

	accepted := false
	for did_somthing := false; !did_somthing; {
		// Choose a random leaf (sub) and leaf's parent (sup).
		sub, sup := m.random_leaf()
		switch {
		case m.random_bit() != 0 && m.is_joinable(sub, sup):
			// Try to join child with parent.
			accepted = sup.join(m)
			did_somthing = true
		case t.n_non_empty_leafs+1 <= float64(max_leafs) &&
			sub.n_pairs() >= m.Min_pairs_for_split:
			// Try to split sub in 2 using random masked bit.
			if bit, ok := sub.random_masked_bit(m); ok {
				accepted = sub.split(m, bit)
			}
			did_somthing = true
		}
	}

	t.n_steps++
	t_last := m.get_tree_seq(m.tree_sequence - 1)
	lower_cost_found = accepted && t.cost < t_last.cost
	if lower_cost_found {
		// Lower cost found: accept tree and advance sequence number.
		if m.tree_sequence > 0 {
			m.free_tree(m.tree_sequence - 1)
		}
		m.tree_sequence++
		t.n_steps = 0
	} else if m.Restart_after_steps != 0 && t.n_steps > m.Restart_after_steps {
		// Lower cost not found after a number of steps: restart and try again.
		m.restart()
		m.stats.n_restart++
	}

	return
}

func (m *Main) Init() {
	for i := range m.trees {
		m.trees[i].Main = m
	}

	m.n_pairs_per_key = m.Key_bits / 32
	if m.Key_bits%32 != 0 {
		m.n_pairs_per_key++
	}
	if m.n_pairs_per_key == 0 {
		panic("no pairs")
	}

	m.pair_hash.init(m.n_pairs_per_key, 0)
	m.temperature = m.Config.Temperature

	if m.wantValidate() {
		m.validate_all_pairs = make(map[maxPair]bool)
	}

	m.new_root(0)

	// Set initial sequence number.  We've initialized sequence 0; we'll optimize tree 1.
	// Tree will be copied on first optimize iteration.
	m.tree_sequence = 1
}

type level struct {
	n_leafs, n_pairs uint
	min_pairs        uint
	max_pairs        uint
}

//go:generate gentemplate -d Package=mctree -id level -d VecType=level_vec -d Type=level github.com/platinasystems/go/elib/vec.tmpl

type tree_stats struct {
	levels         level_vec
	pair_histogram [32]uint32
}

func (ts *tree_stats) count(m *Main, i node_index, l int) {
	n := m.get_node(i)
	if n.is_split() {
		ts.count(m, n.sub_nodes[0], l+1)
		ts.count(m, n.sub_nodes[1], l+1)
	}
	if np := n.n_pairs(); np != 0 {
		ts.levels.Validate(uint(l))
		v := &ts.levels[l]

		j := elib.Word(np).MaxLog2()
		ts.pair_histogram[j]++

		v.n_leafs += 1
		v.n_pairs += np
		if np > v.max_pairs {
			v.max_pairs = np
		}
		if v.min_pairs == 0 {
			v.min_pairs = np
		}
		if np < v.min_pairs {
			v.min_pairs = np
		}
	}
}

func (ts *tree_stats) count_tree(m *Main, t *tree) { ts.count(m, t.root_node_index, 0) }

func (ts *tree_stats) String() (s string) {
	s = ""
	for l := range ts.levels {
		lv := &ts.levels[l]
		if lv.n_leafs > 0 {
			s += fmt.Sprintf("level %2d: %6d leafs %6d pairs ave/min/max %.2f/%d/%d\n",
				l, lv.n_leafs, lv.n_pairs, float64(lv.n_pairs)/float64(lv.n_leafs), lv.min_pairs, lv.max_pairs)
		}
	}

	sum := ts.pair_histogram[0]
	for i := 1; i < len(ts.pair_histogram); i++ {
		sum += ts.pair_histogram[i]
	}
	sum1 := uint32(0)
	for i := len(ts.pair_histogram) - 1; i >= 0; i-- {
		if h := ts.pair_histogram[i]; h > 0 {
			sum1 += h
			n_max := 1 << uint(i)
			n_min := n_max / 2
			s += fmt.Sprintf("  leafs with %3d < pairs <= %3d: %4d %.2f %.2f\n", n_min, n_max, h, float64(h)/float64(sum), float64(sum1)/float64(sum))
		}
	}

	return
}

func (m *Main) Print(i uint, start time.Time, verbose bool) {
	t := m.get_min_tree()
	ts := tree_stats{}
	ts.count_tree(m, t)
	np := float64(m.n_pairs())
	fmt.Printf("iteration %8d: tree sequence %d cost %e leafs %f per leaf %f q %f occupancy %f pairs %d pairvecs\n  join %+v split %+v restarts %d\n  elapsed time: %s\n%s",
		i,
		m.tree_sequence, t.cost, t.n_non_empty_leafs, t.occupancy/t.n_non_empty_leafs,
		t.compute_q(m),
		t.occupancy/np, m.shared_pair_offsets_pool.Elts(),
		&m.stats.join, &m.stats.split, m.stats.n_restart,
		time.Since(start),
		&ts)
	if verbose {
		t.dump(m)
	}
}
