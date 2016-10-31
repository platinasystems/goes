// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tomahawk

import (
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/sbus"

	"fmt"
)

type DmaRequest struct {
	*tomahawk
	sbus.DmaRequest
}

func (r *DmaRequest) Start() {
	r.Dma.Start(&r.DmaRequest)
}

func (r *DmaRequest) Do() {
	r.Start()
	<-r.Done
}

type parallelReq struct {
	reqs   [3]DmaRequest
	active [3]bool
	i, cap int
}

func (p *parallelReq) cur() *DmaRequest { return &p.reqs[p.i] }

func (p *parallelReq) init(t *tomahawk, cap int) *DmaRequest {
	for i := range p.reqs {
		p.reqs[i].tomahawk = t
		p.active[i] = false
	}
	p.i = 0
	p.cap = cap
	return p.cur()
}

func (p *parallelReq) do() (q *DmaRequest) {
	q = p.cur()
	if len(q.Commands) >= p.cap {
		q.Start()
		p.active[p.i] = true
		p.i = p.i + 1
		if p.i >= len(p.reqs) {
			p.i = 0
		}
		q = p.cur()
		if len(q.Commands) > 0 {
			// Synchronous wait for wrapped request still pending.
			<-q.Done
			p.active[p.i] = false
			if q.Err != nil {
				panic(q.Err)
			}
		}
	}
	return
}

func (p *parallelReq) flush() {
	q := &p.reqs[p.i]
	if len(q.Commands) > 0 {
		q.Start()
		p.active[p.i] = true
	}
	for i := range p.reqs {
		q := &p.reqs[i]
		if p.active[i] {
			<-q.Done
			if q.Err != nil {
				panic(q.Err)
			}
			p.active[i] = false
		}
	}
	return
}

const (
	BlockRxPipe    sbus.Block = 1
	BlockTxPipe    sbus.Block = 2
	BlockMmuXpe    sbus.Block = 3
	BlockMmuSc     sbus.Block = 4
	BlockMmuGlobal sbus.Block = 5
	BlockOtpc      sbus.Block = 6
	BlockTop       sbus.Block = 7
	BlockSer       sbus.Block = 8
	BlockCMIC      sbus.Block = 9 // used for schan replies
	BlockIproc     sbus.Block = 10
	BlockXlport0   sbus.Block = 11
	BlockClport0   sbus.Block = 15 // ports 0 - 31
	BlockLoopback0 sbus.Block = 54
	BlockLoopback1 sbus.Block = 51
	BlockLoopback2 sbus.Block = 52
	BlockLoopback3 sbus.Block = 53
	BlockClport32  sbus.Block = 55
	BlockAvs       sbus.Block = 59
)

func sbusBlockString(b sbus.Block) string {
	var n = [...]string{
		BlockRxPipe:    "rx pipe",
		BlockTxPipe:    "tx pipe",
		BlockMmuXpe:    "mmu xpe",
		BlockMmuSc:     "mmu sc",
		BlockMmuGlobal: "mmu global",
		BlockTop:       "top",
		BlockSer:       "ser",
		BlockCMIC:      "cmic",
		BlockAvs:       "avs",
	}
	if int(b) < len(n) && len(n[b]) > 0 {
		return n[b]
	}
	switch {
	case b >= BlockClport0 && b < BlockClport0+32:
		return fmt.Sprintf("clport%d", int(b-BlockClport0))
	case b >= BlockXlport0 && b < BlockXlport0+2:
		return fmt.Sprintf("xlport%d", int(b-BlockXlport0))
	default:
		return fmt.Sprintf("Block(%d)", int(b))
	}
}
func init() { sbus.BlockToString = sbusBlockString }
