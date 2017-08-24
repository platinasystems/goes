// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elib

import (
	"fmt"
	"reflect"
	"sort"
)

func DecodeAddress(x interface{}, address uint) (path []string, t reflect.Type) {
	t = reflect.ValueOf(x).Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	addr := uintptr(address)
	for {
		switch t.Kind() {
		case reflect.Struct:
			nf := t.NumField()
			var nextType reflect.Type
			sort.Search(nf, func(i int) (ok bool) {
				f := t.Field(i)
				lo, hi := f.Offset, f.Offset+f.Type.Size()
				ok = addr <= lo
				if found := addr >= lo && addr < hi; found {
					dot := ""
					if len(path) > 0 {
						dot = "."
					}
					path = append(path, dot+f.Name)
					nextType = f.Type
					addr -= lo
				}
				return
			})
			if nextType == nil {
				panic(fmt.Errorf("not found %s 0x%x %v 0x%x", t.Name(), addr, path, address))
			}
			t = nextType
		case reflect.Array:
			t = t.Elem()
			i0, i1 := addr/t.Size(), addr%t.Size()
			path = append(path, fmt.Sprintf("[%d]", i0))
			addr = i1
		default:
			return
		}
	}
}
