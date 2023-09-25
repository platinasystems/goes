// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import "testing"

func TestHeader(t *testing.T) {
	t.Run("ARP", HeaderCheck[ARP])
	t.Run("ETH", HeaderCheck[ETH])
	t.Run("Hop6", HeaderCheck[Hop6])
	t.Run("ICMP", HeaderCheck[ICMP])
	t.Run("ICMP6", HeaderCheck[ICMP6])
	t.Run("IEEE8021Q", HeaderCheck[IEEE8021Q])
	t.Run("IPv4", HeaderCheck[IPv4])
	t.Run("IPv6", HeaderCheck[IPv6])
	t.Run("MPLS", HeaderCheck[MPLS])
	t.Run("BOS", HeaderCheck[BOS])
	t.Run("TCP", HeaderCheck[TCP])
	t.Run("UDP", HeaderCheck[UDP])
	t.Run("TapPI", HeaderCheck[TapPI])
	t.Run("TunPI", HeaderCheck[TunPI])
}

func HeaderCheck[T Headers](t *testing.T) {
	t.Helper()
	var data [64]byte
	h, rem := Header[T](data[:])
	if got, want := cap(data[:])-len(rem), Sizeof(h); got != want {
		t.Error(got, "!=", want)
	} else {
		t.Log("size:", got)
	}
}
