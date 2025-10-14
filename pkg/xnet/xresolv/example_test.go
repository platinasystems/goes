// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xresolv

import (
	"fmt"
	"strings"
)

const ResolvExamples = `
# This is a comment.
search lan
nameserver 192.168.1.1
sortlist 130.155.160.0/255.255.240.0 130.155.0.0
options ndots:2 edns0 foo:bar
; This is also a comment.
---
# This is empty
---
# This just has nameserver
nameserver 8.8.8.8
---
nameserver 1.1.1.1	# This is a trailing comment.
---
nameserver 8.8.4.4;	This is another trailing comment.
`

func ExampleResolv() {
	HostDomain = func() string { return "foo.bar" }
	sep := "---\n"
	for s := ResolvExamples; len(s) > 0; {
		if i := strings.Index(s, sep); i > 0 {
			fmt.Print(String(s[:i]))
			fmt.Print(sep)
			s = s[i+len(sep):]
		} else {
			fmt.Print(String(s))
			s = s[:0]
		}
	}
	// Output:
	// nameserver 192.168.1.1
	// search lan
	// sortlist 130.155.160.0/255.255.240.0 130.155.0.0
	// options edns0 foo:bar ndots:2
	// ---
	// nameserver localhost
	// search foo.bar
	// ---
	// nameserver 8.8.8.8
	// search foo.bar
	// ---
	// nameserver 1.1.1.1
	// search foo.bar
	// ---
	// nameserver 8.8.4.4
	// search foo.bar
}
