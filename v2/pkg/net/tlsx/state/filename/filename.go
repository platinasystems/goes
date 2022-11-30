// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package filename

import (
	"path/filepath"

	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

type Type = *cache.Cache[string]

var Addresses = cache.New[string](func(p *string) error {
	*p = filename("addresses.json")
	return nil
}).Value

var Cert = cache.New[string](func(p *string) error {
	*p = filename("cert.pem")
	return nil
}).Value

var Subscribers = cache.New[string](func(p *string) error {
	*p = filename("subscribers.pem")
	return nil
}).Value

var Subscriptions = cache.New[string](func(p *string) error {
	*p = filename("subscriptions.pem")
	return nil
}).Value

var PrivateKey = cache.New[string](func(p *string) error {
	*p = filename("key.pem")
	return nil
}).Value

func filename(base string) string {
	return filepath.Join(state.Dir(), base)
}
