// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/cert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
)

func Exchanges() (l []string) {
	if c, err := cert.ValErr(); err == nil {
		l = append(l, c.Leaf.DNSNames[0])
	}
	l = append(l, certs.Subscriptions.Names()...)
	return l
}
