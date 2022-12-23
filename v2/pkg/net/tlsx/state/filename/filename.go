// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package filename

import (
	"path/filepath"

	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/dir"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

type Type = *cache.Cache[string]

var Cert = cache.New[string](func(p *string) error {
	*p = filename("cert.pem")
	return nil
}).Value

var Hosts = cache.New[string](func(p *string) error {
	*p = filename("hosts.json")
	return nil
}).Value

var PrivateKey = cache.New[string](func(p *string) error {
	*p = filename("key.pem")
	return nil
}).Value

var Services = cache.New[string](func(p *string) error {
	*p = filename("services.json")
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

func filename(base string) string {
	return filepath.Join(dir.Name(), base)
}

var ExchangeAddress = cache.New[string](func(p *string) error {
	*p = filename("exchange_address.json")
	return nil
}).Value

var RPCAddress = cache.New[string](func(p *string) error {
	*p = filename("rpc_address.json")
	return nil
}).Value
