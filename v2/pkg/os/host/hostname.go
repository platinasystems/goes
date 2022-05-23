// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package host

import (
	"os"

	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

var Name = cache.New[string](func(p *string) error {
	s, err := os.Hostname()
	if err == nil {
		*p = s
	}
	return err
})

func Set(s string) error {
	return Name.Ref(func(p *string) error {
		err := sethostname([]byte(s))
		if err == nil {
			*p = s
		}
		return err
	})
}
