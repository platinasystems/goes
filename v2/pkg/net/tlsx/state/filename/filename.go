// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package filename

import (
	"os"
	"path/filepath"

	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/dir"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

var Cert = cache.New[string](func(p *string) error {
	*p = filename("TLSX_CERT_FILE", "cert.pem")
	return nil
}).Value

var PrivateKey = cache.New[string](func(p *string) error {
	*p = filename("TLSX_KEY_FILE", "key.pem")
	return nil
}).Value

var Subscribers = cache.New[string](func(p *string) error {
	*p = filename("TLSX_SUBSCRIBERS_FILE", "subscribers.pem")
	return nil
}).Value

var Subscriptions = cache.New[string](func(p *string) error {
	*p = filename("TLSX_SUBSCRIPTIONS_FILE", "subscriptions.pem")
	return nil
}).Value

func filename(env, base string) (fn string) {
	if fn = os.Getenv(env); len(fn) == 0 {
		fn = filepath.Join(dir.Name.Value(), base)
	}
	return
}
