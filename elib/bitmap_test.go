// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elib

import (
	"testing"
)

func TestBitmap(t *testing.T) {
	{
		b := Bitmap(0)
		b = b.Set(64)
		if got, want := b.String(), "{64}"; got != want {
			t.Errorf("Set: got %s want %s", got, want)
		}

		b = b.Free()
		if got, want := b.String(), "{}"; got != want {
			t.Errorf("Free: got %s want %s", got, want)
		}

		b = Bitmap((1 << 3) | (1 << 13))
		b = Bitmaps.Set(b, 128)
		if got, want := b.String(), "{3, 13, 128}"; got != want {
			t.Errorf("Set 128: got %s want %s", got, want)
		}

		c := b.Dup().Set(12)
		if got, want := c.String(), "{3, 12, 13, 128}"; got != want {
			t.Errorf("Dup new: got %s want %s", got, want)
		}
		if got, want := b.String(), "{3, 13, 128}"; got != want {
			t.Errorf("Dup old: got %s want %s", got, want)
		}

		d := c.Dup()
		i := 0
		want := [4]uint{3, 12, 13, 128}
		for x := ^uint(0); d.Next(&x); {
			if x != want[i] {
				t.Errorf("Next %d: got %d want %d", i, x, want[i])
			}
			i++
		}

		c = c.Free()
		d.Free()
		b.Free()
	}

	{
		zero := Bitmap(0)
		c := zero.Set(1).Set(2)
		c = c.Orx(128)
		d := zero.Orx(256)
		if got, want := c.String(), "{1, 2, 128}"; got != want {
			t.Errorf("Orx: got %s want %s", got, want)
		}
		if got, want := d.String(), "{256}"; got != want {
			t.Errorf("Orx: got %s want %s", got, want)
		}

		d = d.Or(c)
		if got, want := d.String(), "{1, 2, 128, 256}"; got != want {
			t.Errorf("Or: got %s want %s", got, want)
		}

		c.Free()

		d = d.AndNotx(256)
		d = d.AndNotx(128)
		if got, want := d.String(), "{1, 2}"; got != want {
			t.Errorf("AndNotx: got %s want %s", got, want)
		}
	}

	{
		var z WordVec
		old := z.SetMultiple(63, 2, 3)
		new := z.GetMultiple(63, 2)
		if old != 0 || new != 3 {
			t.Errorf("SetMultiple old %d new %d", old, new)
		}
	}

	{
		var z Bitmap
		z, old := z.SetMultiple(63, 2, 3)
		new := z.GetMultiple(63, 2)
		if old != 0 || new != 3 {
			t.Errorf("SetMultiple old %d new %d", old, new)
		}
		z.Free()
	}
}
