// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package style

import "github.com/platinasystems/goes/v2/pkg/text/complete"

func Completions(args []string, completers ...any) {
	for _, s := range complete.Last(args, completers...) {
		Plain.Notice.Println(s)
	}
}
