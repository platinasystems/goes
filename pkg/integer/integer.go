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

// *subject += object
func Add[S, O Int | Uint](subject *S, object O) {
	*subject += S(object)
}

// *subject = object
func Assign[S, O Int | Uint](subject *S, object O) {
	*subject = S(object)
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
