// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xflag

import . "flag"

// [FlagSet.Init] with new name but current [FlagSet.ErrorHandling].
func Rename(flags *FlagSet, name string) {
	flags.Init(name, flags.ErrorHandling())
}
