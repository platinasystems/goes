// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package label

import (
	"errors"
	"slices"
	"sync"
)

type Ring struct {
	sync.Mutex
	labels []Label
	next   int
}

var (
	ErrNotFound    = errors.New("not found")
	ErrUnavailable = errors.New("unavailable")
)

func (r *Ring) Append(lbl Label) {
	r.Lock()
	defer r.Unlock()
	r.labels = append(r.labels, lbl)
}

func (r *Ring) Next() (Label, error) {
	r.Lock()
	defer r.Unlock()
	n := len(r.labels)
	if n == 0 {
		return Mislabel, ErrUnavailable
	}
	lbl := r.labels[r.next]
	if r.next += 1; r.next == n {
		r.next = 0
	}
	return lbl, nil
}

func (r *Ring) Remove(lbl Label) error {
	r.Lock()
	defer r.Unlock()
	lbli := lbl.Index()
	for i, ex := range r.labels {
		if ex.Index() == lbli {
			r.labels = slices.Delete(r.labels, i, i+1)
			if r.next >= len(r.labels) {
				r.next = 0
			}
			return nil
		}
	}
	return ErrNotFound
}

func (r *Ring) Update(lbl Label) error {
	r.Lock()
	defer r.Unlock()
	lbli := lbl.Index()
	for i, entry := range r.labels {
		if entry.Index() == lbli {
			r.labels[i] = lbl
			return nil
		}
	}
	return ErrNotFound
}
