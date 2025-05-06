// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xflag

import (
	. "flag"
	"strings"
)

// Returns last space separated [FlagSet.Name].
func LastName(flags *FlagSet) string {
	name := flags.Name()
	if i := strings.LastIndex(name, " "); i > 0 {
		name = name[i+1:]
	}
	return name
}
