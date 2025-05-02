// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bind

// Features map by name these emulated BIND9 tools.
var Features = map[string]any{
	"dig":   Dig,
	"host":  Host,
	"named": Named,

	"named-checkzone": NCZ,
}
