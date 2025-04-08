// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"fmt"
	"slices"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/box"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

type IdRing struct {
	sync.Mutex
	ids  []box.Id
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
		return box.InvalidId, xerrors.Unavailable("id")
	}
	id := r.ids[r.next]
	if r.next += 1; r.next == n {
		r.next = 0
	}
	return id, nil
}

func (r *IdRing) Remove(id box.Id) error {
	r.Lock()
	defer r.Unlock()
	iid := id.Index()
	for i, ex := range r.ids {
		if ex.Index() == iid {
			r.ids = slices.Delete(r.ids, i, i+1)
			if r.next >= len(r.ids) {
				r.next = 0
			}
			return nil
		}
	}
	return xerrors.NotFound("id", fmt.Sprint(id))
}

func (r *IdRing) Update(id box.Id) error {
	r.Lock()
	defer r.Unlock()
	iid := id.Index()
	for i, entry := range r.ids {
		if entry.Index() == iid {
			r.ids[i] = id
			return nil
		}
	}
	return xerrors.NotFound("id", fmt.Sprint(id))
}
