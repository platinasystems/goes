// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"slices"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/box"
)

type IdRing struct {
	sync.Mutex
	ids  []uint32
	next int
}

func (r *IdRing) Append(id box.Id) {
	r.Lock()
	defer r.Unlock()
	r.ids = append(r.ids, id)
}

func (r *IdRing) Next() (box.Id, error) {
	r.Lock()
	defer r.Unlock()
	n := len(r.ids)
	if n == 0 {
		return InvalidId, ErrUnavailable
	}
	id := r.ids[r.next]
	if r.next += 1; r.next == n {
		r.next = 0
	}
	return id, nil
}

func (r *IdRing) Remove(id uint32) error {
	r.Lock()
	defer r.Unlock()
	iid := IdIndex(id)
	for i, ex := range r.ids {
		if IdIndex(ex) == iid {
			r.ids = slices.Delete(r.ids, i, i+1)
			if r.next >= len(r.ids) {
				r.next = 0
			}
			return nil
		}
	}
	return ErrNotFound
}

func (r *IdRing) Update(id uint32) error {
	r.Lock()
	defer r.Unlock()
	iid := IdIndex(id)
	for i, entry := range r.ids {
		if IdIndex(entry) == iid {
			r.ids[i] = id
			return nil
		}
	}
	return ErrNotFound
}
