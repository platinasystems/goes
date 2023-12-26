// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package override

// Override value and return its restoration closure, e.g.
//
//	defer override.Value(&os.Stdout, w)()
func Value[T any](target *T, value T) func() {
	save := *target
	*target = value
	return func() {
		*target = save
	}
}
