// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import "github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"

func Exchanges() (l []string) {
	l = append(l, certs.Self.DNSNames()...)
	l = append(l, certs.Subscriptions.Names()...)
	return l
}
