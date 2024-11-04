// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import "golang.org/x/net/dns/dnsmessage"

type RCode uint16

//go:generate stringer -type RCode -trimprefix RCode -linecomment
const (
	RCodeSuccess RCode = iota // NoError

	RCodeFormatError    // FormErr
	RCodeServerFailure  // ServFail
	RCodeNameError      // NXDomain
	RCodeNotImplemented // NotImp
	RCodeRefused        // Refused
)

func _() {
	// confirm codes
	var x [1]struct{}
	_ = x[int(RCodeSuccess)-int(dnsmessage.RCodeSuccess)]
	_ = x[int(RCodeFormatError)-int(dnsmessage.RCodeFormatError)]
	_ = x[int(RCodeServerFailure)-int(dnsmessage.RCodeServerFailure)]
	_ = x[int(RCodeNameError)-int(dnsmessage.RCodeNameError)]
	_ = x[int(RCodeNotImplemented)-int(dnsmessage.RCodeNotImplemented)]
	_ = x[int(RCodeRefused)-int(dnsmessage.RCodeRefused)]
}
