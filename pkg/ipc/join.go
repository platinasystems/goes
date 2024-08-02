// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import "fmt"

func join(prefix string, args []any) string {
	return fmt.Sprint(prefix, fmt.Sprint(args...))
}
