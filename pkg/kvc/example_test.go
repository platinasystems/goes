// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package kvc

import "fmt"

func Example() {
	RangeString(`
# comment
no comment
// another comment
no-values
; and another comment
with more values
`, nil, func(lno int, k string, v []string) error {
		fmt.Println(k, v)
		return nil
	})
	// Output:
	// no [comment]
	// no-values []
	// with [more values]
}
