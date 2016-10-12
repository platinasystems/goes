package pg

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cpu"
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"
)

const (
	next_error = iota
	next_punt
	n_next
)

const (
	error_none = iota
	tx_packets_dropped
)

type node struct {
	vnet.InterfaceNode
	vnet.HwIf
	v    *vnet.Vnet
	pool vnet.BufferPool
	stream_pool
	stream_index_by_name parse.StringMap
	stream_type_map      parse.StringMap
	stream_types         []StreamType
	buffer_type_pool
	orphan_refs vnet.RefVec
	node_validate
}

const (
	buffer_type_nil = 0xffffffff
)

type buffer_type struct {
	index             uint32
	stream_index      uint
	data_index        uint
	data              []byte
	free_refs         vnet.RefVec
	validate_sequence uint
}

//go:generate gentemplate -d Package=pg -id buffer_type_pool -d PoolType=buffer_type_pool -d Type=buffer_type -d Data=elts github.com/platinasystems/go/elib/pool.tmpl

func (n *node) init(v *vnet.Vnet) {
	n.v = v
	n.Next = []string{
		next_error: "error",
		next_punt:  "punt",
	}
	n.Errors = []string{
		error_none:         "packets generated",
		tx_packets_dropped: "tx packets dropped",
	}
	v.RegisterHwInterface(n, "packet-generator")
	v.RegisterInterfaceNode(n, n.Hi(), "packet-generator")

	// Link is always up for packet generator.
	n.SetLinkUp(true)
	n.SetAdminUp(true)

	p := &n.pool
	t := &p.BufferTemplate
	*t = vnet.DefaultBufferPool.BufferTemplate
	r := p.GetRefTemplate()
	n.SetError(r, error_none)
	t.Buffer.SetSave(buffer_type_nil)
	p.Name = n.Name()
	v.AddBufferPool(p)
}

func (n *node) free_buffer_type(t *buffer_type) {
	if l := t.free_refs.Len(); l > 0 {
		// Mark buffers as no longer being of this type.
		for i := range t.free_refs {
			b := t.free_refs[i].GetBuffer()
			b.SetSave(buffer_type_nil)
		}
		n.pool.FreeRefs(&t.free_refs[0], l, false)
	}
	t.free_refs = t.free_refs[:0]
	t.data = nil
}

func buffer_type_for_size(size, unit uint) (n uint) {
	for size > unit {
		n++
		size -= unit
	}
	return
}

func (n *node) setData(s *Stream) {
	// Return cached refs in pool.
	n.return_buffers()

	// Free previously used buffer types.
	for _, t := range s.buffer_types {
		n.free_buffer_type(&n.buffer_type_pool.elts[t])
		n.buffer_type_pool.PutIndex(uint(t))
	}

	n_data := uint(len(s.data))
	size := n.pool.BufferTemplate.Size
	n_size := 1 + buffer_type_for_size(n_data, size)
	s.buffer_types.Validate(n_size - 1)
	s.buffer_types = s.buffer_types[:n_size]
	i, j := uint(0), uint(0)
	for i < n_data {
		this_size := size
		if i+this_size > n_data {
			this_size = n_data - i
		}
		bi := uint32(n.buffer_type_pool.GetIndex())
		s.buffer_types[j] = bi
		t := &n.buffer_type_pool.elts[bi]
		t.index = bi
		t.stream_index = s.index
		t.data_index = j
		t.data = s.data[i : i+this_size]
		j++
		i += this_size
	}
}

func (n *node) free1(r0 *vnet.Ref, ti0 uint32) {
	t0 := &n.buffer_type_pool.elts[ti0]
	r0.SetDataLen(uint(len(t0.data)))
	t0.validate_ref(r0)
	t0.free_refs = append(t0.free_refs, *r0)
}

func (n *node) free_buffers(refs vnet.RefVec, t *buffer_type) {
	i, n_left := uint(0), refs.Len()

	fl := t.free_refs.Len()
	tf := t.free_refs
	tf.Resize(n_left)
	ti := t.index
	fi := fl
	tl := uint(len(t.data))

	for n_left >= 4 {
		r0, r1, r2, r3 := &refs[i+0], &refs[i+1], &refs[i+2], &refs[i+3]
		b0, b1, b2, b3 := r0.GetBuffer(), r1.GetBuffer(), r2.GetBuffer(), r3.GetBuffer()
		ti0, ti1, ti2, ti3 := uint32(b0.GetSave()), uint32(b1.GetSave()), uint32(b2.GetSave()), uint32(b3.GetSave())
		r0.SetDataLen(tl)
		r1.SetDataLen(tl)
		r2.SetDataLen(tl)
		r3.SetDataLen(tl)
		tf[fi+0] = *r0
		tf[fi+1] = *r1
		tf[fi+2] = *r2
		tf[fi+3] = *r3
		i += 4
		fi += 4
		n_left -= 4
		if ti0 != ti || ti1 != ti || ti2 != ti || ti3 != ti {
			fi -= 4
			fi = n.slow_path(t, tf, fi, r0, b0, ti0)
			fi = n.slow_path(t, tf, fi, r1, b1, ti1)
			fi = n.slow_path(t, tf, fi, r2, b2, ti2)
			fi = n.slow_path(t, tf, fi, r3, b3, ti3)
			tf.ValidateLen(fi + n_left)
		}
	}

	for n_left > 0 {
		r0 := &refs[i]
		b0 := r0.GetBuffer()
		ti0 := uint32(b0.GetSave())
		r0.SetDataLen(tl)
		tf[fi+0] = *r0
		i += 1
		fi += 1
		n_left -= 1
		if ti0 != ti {
			fi -= 1
			fi = n.slow_path(t, tf, fi, r0, b0, ti0)
			tf.ValidateLen(fi + n_left)
		}
	}

	t.free_refs = tf[:fi]
}

func (n *node) slow_path(t *buffer_type, tf []vnet.Ref, fiʹ uint, r0 *vnet.Ref, b0 *vnet.Buffer, ti0 uint32) (fi uint) {
	fi = fiʹ
	if ti0 == buffer_type_nil {
		ti0 = t.index
		r0.SetDataLen(uint(len(t.data)))
		copy(r0.DataSlice(), t.data)
		b0.SetSave(hw.BufferSave(ti0))
	}

	t0 := &n.buffer_type_pool.elts[ti0]
	var f0 *vnet.Ref
	if t0 == t {
		f0 = &tf[fi]
		fi++
	} else {
		l := t0.free_refs.Len()
		t0.free_refs.Resize(1)
		f0 = &t0.free_refs[l]
	}

	*f0 = *r0
	f0.SetDataLen(uint(len(t0.data)))
	t0.validate_ref(f0)
	return
}

func (n *node) return_buffers() {
	refs := n.pool.AllocCachedRefs()
	for i := range refs {
		r0 := &refs[i]
		b0 := r0.GetBuffer()
		ti0 := uint32(b0.GetSave())
		if ti0 == buffer_type_nil {
			n.orphan_refs = append(n.orphan_refs, *r0)
		} else {
			n.free1(r0, ti0)
		}
	}
	if l := n.orphan_refs.Len(); l > 0 {
		n.pool.FreeRefs(&n.orphan_refs[0], l, false)
		n.orphan_refs = n.orphan_refs[:0]
	}
}

func (n *node) buffer_type_get_refs(dst []vnet.Ref, want, ti uint) {
	t := &n.buffer_type_pool.elts[ti]
	var got uint
	for {
		if got = t.free_refs.Len(); got >= want {
			break
		}
		var tmp [vnet.MaxVectorLen]vnet.Ref
		n.pool.AllocRefs(tmp[:])
		n.free_buffers(tmp[:], t)
	}

	copy(dst, t.free_refs[got-want:got])

	if elib.Debug {
		for i := uint(0); i < want; i++ {
			t.validate_ref(&dst[i])
		}
	}

	t.free_refs = t.free_refs[:got-want]
	return
}

type node_validate struct {
	validate_data     []byte
	validate_sequence uint
}

func (n *node) generate_n_types(s *Stream, dst []vnet.Ref, n_packets, n_types uint) (n_bytes uint) {
	var tmp [4][vnet.MaxVectorLen]vnet.Ref
	var prev, prev_prev []vnet.Ref
	this := dst
	save := s.cur_size
	is_single_size := s.max_size == s.min_size
	d := (n_types - 1) * n.pool.Size
	n_bytes = d * n_packets
	if is_single_size {
		n_bytes = s.min_size * n_packets
	}
	for i := uint(0); i < n_types; i++ {
		n.buffer_type_get_refs(this, n_packets, uint(s.buffer_types[i]))
		if i+1 >= n_types && !is_single_size {
			for j := uint(0); j < n_packets; j++ {
				last_size := s.cur_size - d
				this[j].SetDataLen(last_size)
				n_bytes += last_size
				s.cur_size = s.next_size(s.cur_size, 0)
			}
		}
		if prev != nil {
			var pp *hw.RefHeader
			if prev_prev != nil {
				pp = &prev_prev[0].RefHeader
			}
			hw.LinkRefs(pp, &prev[0].RefHeader, &this[0].RefHeader, 1+i, n_packets)
		}
		prev_prev = prev
		prev = this
		this = tmp[i&3][:]
	}

	if elib.Debug {
		save, s.cur_size = s.cur_size, save
		for i := uint(0); i < n_packets; i++ {
			n.validate_ref(&dst[i], s)
			s.cur_size = s.next_size(s.cur_size, 0)
		}
		s.cur_size = save
	}

	return
}

func (n *node) generate(s *Stream, dst []vnet.Ref, n_packets uint) (n_bytes uint) {
	n_left := n_packets
	for {
		nt := 1 + buffer_type_for_size(s.cur_size, n.pool.Size)
		n_this := n_left
		if s.max_size != s.min_size {
			n_this = 1 + s.max_size - s.cur_size
			if next := 1 + nt*n.pool.Size - s.cur_size; n_this > next {
				n_this = next
			}
			if n_this > n_left {
				n_this = n_left
			}
		}
		n_bytes += n.generate_n_types(s, dst[n_packets-n_left:], n_this, nt)
		n_left -= n_this
		if n_left == 0 {
			break
		}
	}
	return
}

func (n *node) n_packets_this_input(s *Stream, cap uint) (p uint, dt_next float64) {
	if s.n_packets_limit == 0 { // unlimited
		p = cap
	} else if s.n_packets_sent < s.n_packets_limit {
		max := s.n_packets_limit - s.n_packets_sent
		if max > uint64(cap) {
			p = cap
		} else {
			p = uint(max)
		}
	}
	if s.rate_packets_per_sec != 0 {
		now := cpu.TimeNow()
		dt := n.Vnet.TimeDiff(now, s.last_time)
		s.credit_packets += dt * s.rate_packets_per_sec
		if float64(p) > s.credit_packets {
			p = uint(s.credit_packets)
		}
		s.credit_packets -= float64(p)
		s.last_time = now
		if s.credit_packets < 1 {
			dt_next = (1 - s.credit_packets) / s.rate_packets_per_sec
		}
	}
	return
}

func (n *node) stream_input(o *vnet.RefOut, s *Stream) (done bool, dt float64) {
	out := &o.Outs[s.next]
	out.BufferPool = &n.pool
	t := n.GetIfThread()

	var n_packets uint
	n_packets, dt = n.n_packets_this_input(s, out.Cap())
	if n_packets > 0 {
		n_bytes := n.generate(s, out.Refs[:], n_packets)
		vnet.IfRxCounter.Add(t, n.Si(), n_packets, n_bytes)
		out.SetPoolAndLen(n.Vnet, &n.pool, n_packets)
		s.n_packets_sent += uint64(n_packets)
	}
	done = s.n_packets_limit != 0 && s.n_packets_sent >= s.n_packets_limit
	return
}

func (n *node) InterfaceInput(o *vnet.RefOut) {
	all_done := true
	min_dt, min_dt_valid := float64(0), false
	n.stream_pool.Foreach(func(s Streamer) {
		done, dt := n.stream_input(o, s.get_stream())
		all_done = all_done && done
		if dt > 0 {
			if !min_dt_valid {
				min_dt = dt
				min_dt_valid = true
			} else if dt < min_dt {
				min_dt = dt
			}
		}
	})
	// For any wait "too small" (10 microseconds), just remain active.
	if !all_done && min_dt > 10e-6 {
		n.ActivateAfterTime(min_dt)
	} else {
		n.Activate(!all_done)
	}
}

func (n *node) InterfaceOutput(i *vnet.TxRefVecIn) {
	n.CountError(tx_packets_dropped, i.NPackets())
	n.Vnet.FreeTxRefIn(i)
}
