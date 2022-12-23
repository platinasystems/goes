// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package host

import (
	"os"

	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

type CachedName struct{ *cache.Cache[string] }

var Name = CachedName{cache.New[string](func(p *string) error {
	s, err := os.Hostname()
	if err == nil {
		*p = s
	}
	return err
})}

func (cn CachedName) UnmarshalText(text []byte) error {
	return cn.Mutex(func(p *string) error {
		err := sethostname(text)
		if err == nil {
			*p = string(text)
		}
		return err
	})
}
