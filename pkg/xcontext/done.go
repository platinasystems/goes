// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xcontext

import . "context"

func IsDone(ctx Context) (ok bool) {
	select {
	case <-ctx.Done():
		ok = true
	default:
	}
	return
}
