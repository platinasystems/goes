// Copyright © 2015-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package align

type Align int

func (to Align) Roundup(i int) int {
	return (i + to.Size() - 1) & ^(to.Size() - 1)
}

func (to Align) Size() int { return int(to) }
