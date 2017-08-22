// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/loop"
)

type interfaceInputer interface {
	InterfaceInput(out *RefOut)
}

type outputInterfaceNoder interface {
	Noder
	GetInterfaceNode() *interfaceNode
	InterfaceOutput(in *TxRefVecIn)
}

type inputOutputInterfaceNoder interface {
	outputInterfaceNoder
	interfaceInputer
}

type interfaceNode struct {
	Node

	hi Hi

	max_tx_refs uint
	cur_tx_refs uint
	tx_chan     chan *TxRefVecIn
	free_list   []*TxRefVecIn

	txDownDropError uint

	freeChan chan *TxRefVecIn
	tx       outputInterfaceNoder
	rx       interfaceInputer
}

type OutputInterfaceNode struct{ interfaceNode }
type InterfaceNode struct{ interfaceNode }

func (n *interfaceNode) MakeLoopIn() loop.LooperIn                { return &RefIn{} }
func (n *interfaceNode) MakeLoopOut() loop.LooperOut              { return &RefOut{} }
func (n *interfaceNode) LoopOutput(l *loop.Loop, i loop.LooperIn) { n.ifOutput(i.(*RefIn)) }
func (n *interfaceNode) GetInterfaceNode() *interfaceNode         { return n }

func (n *InterfaceNode) LoopInput(l *loop.Loop, o loop.LooperOut) {
	n.rx.InterfaceInput(o.(*RefOut))
}

func (v *Vnet) registerInterfaceNodeHelper(n outputInterfaceNoder, hi Hi) {
	x := n.GetInterfaceNode()
	x.txDownDropError = uint(len(x.Errors))
	x.Errors = append(x.Errors, "tx down drops")
	x.hi = hi
	x.setupTx(n)
	h := v.HwIf(hi)
	h.n = append(h.n, n)
}

func (v *Vnet) RegisterOutputInterfaceNode(n outputInterfaceNoder, hi Hi, name string, args ...interface{}) {
	v.registerInterfaceNodeHelper(n, hi)
	v.RegisterNode(n, name, args...)
}

func (v *Vnet) RegisterInterfaceNode(n inputOutputInterfaceNoder, hi Hi, name string, args ...interface{}) {
	x := n.GetInterfaceNode()
	x.rx = n
	v.registerInterfaceNodeHelper(n, hi)
	v.RegisterNode(n, name, args...)
}

const tx_ref_vec_in_fifo_len = 256

func (n *interfaceNode) ifOutputThread() {
	for x := range n.tx_chan {
		n.tx.InterfaceOutput(x)
	}
}

func (n *interfaceNode) setupTx(tx outputInterfaceNoder) {
	n.tx = tx
	n.freeChan = make(chan *TxRefVecIn, tx_ref_vec_in_fifo_len)
	n.max_tx_refs = 2 * MaxVectorLen
}

func (h *HwIf) txNodeUpDown(isUp bool) {
	for _, hn := range h.n {
		n := hn.GetInterfaceNode()
		if isUp {
			n.tx_chan = make(chan *TxRefVecIn, tx_ref_vec_in_fifo_len)
			go n.ifOutputThread()
		} else {
			close(n.tx_chan)
			n.tx_chan = nil
		}
	}
}

type txElogEvent struct {
	node_name elog.StringRef
	n_refs    uint32
}

func (e *txElogEvent) SetData(x *elog.Context, p elog.Pointer) { *(*txElogEvent)(p) = *e }
func (e *txElogEvent) Format(x *elog.Context, f elog.Format) string {
	return f("tx %s send %d buffers", x.GetString(e.node_name), e.n_refs)
}

func (n *interfaceNode) send(i *TxRefVecIn) {
	if elog.Enabled() {
		e := txElogEvent{
			node_name: n.ElogName(),
			n_refs:    uint32(i.Len()),
		}
		elog.Add(&e)
	}
	n.cur_tx_refs += i.Len()
	n.tx_chan <- i
}

func (n *interfaceNode) allocTxRefVecIn(in *RefIn) (i *TxRefVecIn) {
	l := &n.Vnet.loop
	for {
		// Find a place to put a vector of packets (TxRefVecIn).
		select {
		case i = <-n.freeChan:
			// Re-cycle one that output routine is done with.
			i.FreeRefs(false)
			n.cur_tx_refs -= i.Len()
		default:
			if l := len(n.free_list); l > 0 {
				// Re-cycle one from free list.
				i = n.free_list[l-1]
				n.free_list = n.free_list[:l-1]
			} else {
				// Make a new one.
				i = &TxRefVecIn{n: n}
			}
			i.refInCommon = in.refInCommon
			i.nPackets = 0
		}
		// Check that we stay within ref limit.
		if n.cur_tx_refs+MaxVectorLen <= n.max_tx_refs {
			return
		}

		// We're over ref limit.  Add ref back to free list.
		n.free_list = append(n.free_list, i)

		// Suspend unless another has appeared.
		if len(n.freeChan) == 0 {
			l.Suspend(&in.In)
		}
	}
}

func (n *interfaceNode) newTxRefVecIn(in *RefIn, r []Ref) (i *TxRefVecIn) {
	i = n.allocTxRefVecIn(in)
	l := uint(len(r))
	if l > 0 {
		i.Refs.Validate(l - 1)
		copy(i.Refs[0:], r)
	}
	i.Refs = i.Refs[:l]
	return
}

type TxRefVecIn struct {
	RefVecIn
	n *interfaceNode
}

func (v *Vnet) FreeTxRefIn(i *TxRefVecIn) {
	i.n.freeChan <- i
	v.loop.Resume(&i.In)
}
func (i *TxRefVecIn) Free(v *Vnet) { v.FreeTxRefIn(i) }

func (n *interfaceNode) ifOutput(ri *RefIn) {
	if n.tx_chan == nil {
		l := ri.InLen()
		n.CountError(n.txDownDropError, l)
		ri.FreeRefs(l)
		return
	}

	rvi := n.allocTxRefVecIn(ri)
	n_packets_in := ri.InLen()

	// Copy common fields.
	rvi.refInCommon = ri.refInCommon

	rvi.Refs.Validate(n_packets_in - 1)
	rvi.Refs = rvi.Refs[:n_packets_in]

	// Number of packets left to process.
	n_ref_left := n_packets_in

	rs := ri.Refs[:]
	rv := rvi.Refs
	is, iv := uint(0), uint(0)
	n_bytes_in, n_packets_rvi := uint(0), uint(0)
	for n_ref_left >= 4 {
		rv[iv+0] = rs[is+0]
		rv[iv+1] = rs[is+1]
		rv[iv+2] = rs[is+2]
		rv[iv+3] = rs[is+3]
		n_bytes_in += rs[is+0].DataLen() + rs[is+1].DataLen() + rs[is+2].DataLen() + rs[is+3].DataLen()
		iv += 4
		is += 4
		n_ref_left -= 4
		n_packets_rvi += 4
		if RefFlag4(NextValid, rs, is-4) || iv > MaxVectorLen {
			iv -= 4
			n_packets_rvi -= 4
			rvi, rv, iv, n_bytes_in, n_packets_rvi = n.slowPath(ri, rvi, rv, rs, is-4, iv, n_bytes_in, n_packets_rvi)
			rvi, rv, iv, n_bytes_in, n_packets_rvi = n.slowPath(ri, rvi, rv, rs, is-3, iv, n_bytes_in, n_packets_rvi)
			rvi, rv, iv, n_bytes_in, n_packets_rvi = n.slowPath(ri, rvi, rv, rs, is-2, iv, n_bytes_in, n_packets_rvi)
			rvi, rv, iv, n_bytes_in, n_packets_rvi = n.slowPath(ri, rvi, rv, rs, is-1, iv, n_bytes_in, n_packets_rvi)
			rv.ValidateLen(iv + n_ref_left)
		}
	}
	rv.ValidateLen(iv + n_ref_left)
	for n_ref_left > 0 {
		rv[iv+0] = rs[is+0]
		n_bytes_in += rs[is+0].DataLen()
		is += 1
		iv += 1
		n_ref_left -= 1
		n_packets_rvi += 1
		if RefFlag1(NextValid, rs, is-1) || iv > MaxVectorLen {
			iv -= 1
			n_packets_rvi -= 1
			rvi, rv, iv, n_bytes_in, n_packets_rvi = n.slowPath(ri, rvi, rv, rs, is-1, iv, n_bytes_in, n_packets_rvi)
			rv.ValidateLen(iv + n_ref_left)
		}
	}

	if iv > MaxVectorLen {
		panic("overflow")
	}

	// Bump interface packet and byte counters.
	t := n.Vnet.GetIfThread(ri.ThreadId())
	hw := n.Vnet.HwIf(n.hi)
	IfTxCounter.Add(t, hw.si, n_packets_in, n_bytes_in)

	if iv > 0 {
		rvi.Refs = rv[:iv]
		rvi.nPackets = n_packets_rvi

		// Send to output thread, which then calls n.tx.InterfaceOutput.
		n.send(rvi)
	} else {
		n.freeChan <- rvi
	}
}

// Slow path: copy whole packet (not just first ref) to vector.
func (n *interfaceNode) slowPath(
	ri *RefIn, rviʹ *TxRefVecIn, rvʹ RefVec, rs []Ref, is, ivʹ, n_bytesʹ, n_packetsʹ uint) (
	rvi *TxRefVecIn, rv RefVec, iv, n_bytes, n_packets uint) {
	rvi, rv, iv, n_bytes, n_packets = rviʹ, rvʹ, ivʹ, n_bytesʹ, n_packetsʹ
	s := rs[is]

	n_packets++
	for {
		// Copy buffer reference.
		rv.Validate(iv)
		rv[iv] = s
		iv++

		if h := s.RefHeader.NextRef(); h == nil {
			break
		} else {
			s.RefHeader = *h
		}
		n_bytes += s.DataLen()
	}

	// Tx ref vector must not exceed vector length; also, it must contain only full packets.
	// Enfoce this.
	if iv >= MaxVectorLen {
		var save [MaxVectorLen]Ref
		n_save := uint(0)
		if iv > MaxVectorLen {
			n_save = iv - ivʹ

			// Packet must fit into a single vector.
			if n_save > MaxVectorLen {
				panic("packet too large")
			}

			copy(save[:n_save], rv[iv-n_save:iv])
			rv = rv[:ivʹ]
			n_packets--
		} else {
			// Last packet exactly fits.
			rv = rv[:iv]
		}

		// Output current vector and get a new one (possibly suspending).
		rvi.Refs = rv
		rvi.nPackets = n_packets
		n.send(rvi)
		rvi = n.newTxRefVecIn(ri, save[:n_save])
		rv = rvi.Refs
		iv = n_save
		n_packets = 0
		if n_save > 0 {
			n_packets = 1
		}
	}
	return
}

// Transmit ring common code.
type TxDmaRing struct {
	v           *Vnet
	ToInterrupt chan *TxRefVecIn
	o           *TxRefVecIn
	n           uint
}

func (r *TxDmaRing) Init(v *Vnet) {
	r.v = v
	r.ToInterrupt = make(chan *TxRefVecIn, tx_ref_vec_in_fifo_len)
}

func (r *TxDmaRing) InterruptAdvance(n uint) {
	for n > 0 {
		// Nothing in current output vector: refill from channel.
		if r.n == 0 {
			r.o = <-r.ToInterrupt
			r.n = r.o.Len()
		}

		// Advanced past end of current output vector?
		if n < r.n {
			r.n -= n
			break
		}

		// If so, free it.
		n -= r.n
		r.n = 0
		r.o.Free(r.v)
	}
}
