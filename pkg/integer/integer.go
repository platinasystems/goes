// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

// This package provides operators for mismatched types of integers.
package integer

import (
	"fmt"
	"strings"
)

type Int interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Uint interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type UintStringer interface {
	Uint
	fmt.Stringer
}

// *subject += object
func Add[S, O Int | Uint](subject *S, object O) {
	*subject += S(object)
}

// *subject = object
func Assign[S, O Int | Uint](subject *S, object O) {
	*subject = S(object)
}

// Set the named bits that match the indixed names created
// by [golang.org/x/tools/cmd/stringer].
func Bits[V Int | Uint, I Int | Uint](
	base V, named []string, names string, indices ...I,
) (
	v V, found bool,
) {
	v = base
	names = strings.ToLower(names)
	for _, name := range named {
		name = strings.ToLower(name)
		for i := 0; !found && i < len(indices)-1; i++ {
			found = name == names[indices[i]:indices[i+1]]
			if found {
				v |= 1 << i
			}
		}
	}
	return
}

// This returns the comma separated names from LSB to MSB of the subject's true
// bits.
func BitStrings[V Int | Uint, B UintStringer](v V, begin, end B) string {
	const space = " "
	var sep string
	w := new(strings.Builder)
	for bit := begin; bit < end; bit++ {
		if (v & V(1<<bit)) != 0 {
			fmt.Fprint(w, sep, bit)
			sep = space
		}
	}
	return w.String()
}

// int64(subject) - int64(object)
func Diff[S, O Int | Uint](subject S, object O) int64 {
	return int64(subject) - int64(object)
}

// subject == object
func Equal[S, O Int | Uint](subject S, object O) bool {
	return subject == S(object)
}

// (subject & object) == object
func Has[S, O Int | Uint](subject S, object O) bool {
	o := S(object)
	return (subject & o) == o
}

// subject &^ object
func Mask[S, O Int | Uint](subject S, object O) S {
	return subject &^ S(object)
}

// If the subject is within range, the returned name is names[subject],
// otherwise, it's Sprint(subject).
func Name[S Int | Uint](subject S, names []string) string {
	if i := int(subject); 0 <= i && i < len(names) {
		return names[i]
	}
	return fmt.Sprint(subject)
}

// Return the named value that matches one of the indixed names created
// by [golang.org/x/tools/cmd/stringer].
func Named[V Int | Uint, I Int | Uint](
	base V, named, names string, indices ...I,
) (
	v V, found bool,
) {
	named = strings.ToLower(named)
	names = strings.ToLower(names)
	for i := 0; !found && i < len(indices)-1; i++ {
		if found = named == names[indices[i]:indices[i+1]]; found {
			v = base + V(i)
		}
	}
	return
}

// This returns the comma separated names from LSB to MSB of the subject's true
// bits.
func Names[S Int | Uint](subject S, names []string) string {
	var comma string
	w := new(strings.Builder)
	for i, name := range names {
		if (subject & (1 << i)) != 0 {
			fmt.Fprint(w, comma, name)
			comma = ","
		}
	}
	return w.String()
}

// *subject &^= object
func Reset[S, O Int | Uint](subject *S, object O) {
	*subject &^= S(object)
}

// *subject |= S(object)
func Set[S, O Int | Uint](subject *S, object O) {
	*subject |= S(object)
}

// *subject -= object
func Sub[S, O Int | Uint](subject *S, object O) {
	*subject -= S(object)
}

// uint64(subject) + uint64(object)
func Sum[S, O Int | Uint](subject S, object O) uint64 {
	return uint64(subject) + uint64(object)
}

// *subject ^= object
func Toggle[S, O Int | Uint](subject *S, object O) {
	*subject ^= S(object)
}
