// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"fmt"
	"strings"
)

// ErrorfWith context path preface.
func ErrorfWith(ctx context.Context, format string, args ...interface{}) error {
	sb := new(strings.Builder)
	for i, s := range PathOf(ctx) {
		if i > 0 {
			fmt.Fprint(sb, " ")
		}
		fmt.Fprint(sb, s)
	}
	fmt.Fprint(sb, ": ", format)
	return fmt.Errorf(sb.String(), args...)
}
