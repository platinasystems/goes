// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package emptych

func Example() {
	in, out := New()
	go func(in In) {
		close(in)
	}(in)
	<-out
	// Output:
}
