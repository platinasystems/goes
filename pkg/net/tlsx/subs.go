// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"sync"

	"github.com/platinasystems/goes/v2/pkg/crypto/xcert"
)

var Subscribers = sync.
	OnceValue(xcert.NameX509File(SubscribersFileName).Load)
var Subscriptions = sync.
	OnceValue(xcert.NameX509File(SubscriptionsFileName).Load)

func Match(nameOrSKI string) (x *xcert.X509) {
	if self := Self(); self.IsMatch(nameOrSKI) {
		x = &self.X509
	} else if x = Subscriptions().Match(nameOrSKI); x == nil {
		x = Subscribers().Match(nameOrSKI)
	}
	return
}
