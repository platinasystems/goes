// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"github.com/platinasystems/goes/v2/pkg/integer"
	"golang.org/x/net/dns/dnsmessage"
)

// Header Flag Bits
//
//go:generate stringer -type HFB -trimprefix HFB -linecomment
type HFB uint8

const (
	HFBResponse HFB = iota // qr

	HFBAuthoritative      // aa
	HFBTruncated          // tr
	HFBRecursionDesired   // rd
	HFBRecursionAvailable // ra
	HFBAuthenticData      // ad
	HFBCheckingDisabled   // cd

	HFBend
)

// Header Flags
type HF uint8

const (
	HFResponse           HF = 1 << HFBResponse
	HFAuthoritative      HF = 1 << HFBAuthoritative
	HFTruncated          HF = 1 << HFBTruncated
	HFRecursionDesired   HF = 1 << HFBRecursionDesired
	HFRecursionAvailable HF = 1 << HFBRecursionAvailable
	HFAuthenticData      HF = 1 << HFBAuthenticData
	HFCheckingDisabled   HF = 1 << HFBCheckingDisabled
)

func NewHeaderFlags(h dnsmessage.Header) (hf HF) {
	if h.Response {
		hf |= HFResponse
	}
	if h.Authoritative {
		hf |= HFAuthoritative
	}
	if h.Truncated {
		hf |= HFTruncated
	}
	if h.RecursionDesired {
		hf |= HFRecursionDesired
	}
	if h.RecursionAvailable {
		hf |= HFRecursionAvailable
	}
	if h.AuthenticData {
		hf |= HFAuthenticData
	}
	if h.CheckingDisabled {
		hf |= HFCheckingDisabled
	}
	return
}

func (hf HF) String() string {
	return integer.BitStrings(hf, HFB(0), HFBend)
}
