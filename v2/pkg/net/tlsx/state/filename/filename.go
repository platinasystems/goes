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

var Aliases = cache.New[string](func(p *string) (err error) {
	*p, err = filename("aliases.json")
	return
})

var Authorized = cache.New[string](func(p *string) (err error) {
	*p, err = filename("authorized.txt")
	return
})

var Cert = cache.New[string](func(p *string) (err error) {
	*p, err = filename("cert.pem")
	return
})

var Clients = cache.New[string](func(p *string) (err error) {
	*p, err = filename("clients.pem")
	return
})

var Exchanges = cache.New[string](func(p *string) (err error) {
	*p, err = filename("exchanges.pem")
	return
})

var PrivateKey = cache.New[string](func(p *string) (err error) {
	*p, err = filename("key.pem")
	return
})

func filename(base string) (string, error) {
	fn, err := state.Dir.ValErr()
	if err == nil {
		fn = filepath.Join(fn, base)
	}
	return fn, err
}
