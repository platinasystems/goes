// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bind

import (
	"sync"

	"github.com/platinasystems/goes/v2/pkg/xmain"
)

var Version = sync.OnceValue(func() string {
	ver := "(unavailable)"
	if mm := xmain.Module(); mm != nil {
		ver = mm.Version
	}
	return ver
})
