// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package selctx

import (
	"fmt"
	"sort"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context"
)

var Parameter = context.NewParameter(map[string]any{})

func StringIn(ctx context.Context) string {
	m := Parameter.In(ctx)
	keys := make([]string, 0, len(m))
	for k := range m {
		if k != "daemon" && !strings.HasPrefix(k, "_") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		fmt.Fprintln(&sb, " ", k)
	}
	return sb.String()
}
