// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tomahawk

import (
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/sbus"
)

type exact_match_key_type uint8

const (
	key_128 exact_match_key_type = iota
	key_160
	key_320
)

type exact_match_entry struct {
	hit               bool
	valid             bool
	key_type          exact_match_key_type
	qos_profile_id    uint8
	action_profile_id uint8
	class_id          uint16
	action_data       uint64

	// 128 or 160 bit key depending on key type.
	key [5]uint32
}

func (r *exact_match_entry) MemBits() int { return 211 }
func (r *exact_match_entry) MemGetSet(b []uint32, isSet bool) {
	var kt [2]uint8
	var key [3]uint64
	var valid [2]bool

	m.MemGetSet1(&r.hit, b, 210, isSet)
	m.MemGetSet(&r.action_data, b, 209, 192, isSet)
	m.MemGetSetUint8(&r.action_profile_id, b, 191, 187, isSet)
	m.MemGetSetUint8(&r.qos_profile_id, b, 186, 180, isSet)
	m.MemGetSetUint16(&r.class_id, b, 179, 168, isSet)

	if isSet {
		key[0] = m.MemGet(r.key[:], 63, 0)
		key[1] = m.MemGet(r.key[:], 100, 64)
		key[2] = m.MemGet(r.key[:], 159, 101)
		for i := range kt {
			kt[i] = uint8(r.key_type)
			valid[i] = r.valid
		}
	}

	m.MemGetSet(&key[2], b, 167, 109, isSet)
	m.MemGetSetUint8(&kt[1], b, 108, 107, isSet)
	m.MemGetSet1(&valid[1], b, 106, isSet)
	m.MemGetSet(&key[1], b, 104, 68, isSet)
	m.MemGetSet(&key[0], b, 67, 4, isSet)
	m.MemGetSetUint8(&kt[0], b, 3, 2, isSet)
	m.MemGetSet1(&valid[0], b, 1, isSet)

	if !isSet {
		if valid[0] != valid[1] {
			panic("valid bit mismatch")
		}
		if kt[0] != kt[1] {
			panic("key type mismatch")
		}
		r.key_type = exact_match_key_type(kt[0])
		r.valid = valid[0]

		var k [5]uint32
		m.MemSet(k[:], 63, 0, key[0])
		m.MemSet(k[:], 100, 64, key[1])
		m.MemSet(k[:], 159, 101, key[2])
		r.key = k
	}
}

type exact_match_mem m.MemElt

func (r *exact_match_mem) geta(q *DmaRequest, v *exact_match_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *exact_match_mem) seta(q *DmaRequest, v *exact_match_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *exact_match_mem) get(q *DmaRequest, v *exact_match_entry) { r.geta(q, v, sbus.Duplicate) }
func (r *exact_match_mem) set(q *DmaRequest, v *exact_match_entry) { r.seta(q, v, sbus.Duplicate) }
