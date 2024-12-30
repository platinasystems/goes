// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_conf

import (
	"net/netip"
	"testing"
	"time"
)

func CoreValueTest[T CoreValue](
	t *testing.T,
	input string,
	f func(string) (T, error),
	expect T,
) {
	t.Helper()
	t.Run(input, func(t *testing.T) {
		t.Helper()
		if have, err := f(input); err != nil {
			t.Error(err)
		} else if have != expect {
			t.Error("MISMATCH:", have)
		}
	})
}

func CoreValueError[T CoreValue](
	t *testing.T,
	input string,
	f func(string) (T, error),
) {
	t.Helper()
	t.Run(input, func(t *testing.T) {
		t.Helper()
		if have, err := f(input); err == nil {
			t.Error("ESCAPE:", have)
		} else {
			t.Log("CAUGHT:", err)
		}
	})
}

func TestCoreValue(t *testing.T) {
	t.Run("Boolean", func(t *testing.T) {
		for _, test := range []struct {
			s string
			x bool
		}{
			{"0", false},
			{"no", false},
			{"false", false},
			{"1", true},
			{"yes", true},
			{"true", true},
		} {
			CoreValueTest(t, test.s, ParseBool, test.x)
		}
		t.Run("Error", func(t *testing.T) {
			for _, s := range []string{
				"foo",
				"2",
			} {
				CoreValueError(t, s, ParseBool)
			}
		})
	})
	t.Run("Duration", func(t *testing.T) {
		for _, test := range []struct {
			s string
			x time.Duration
		}{
			{"10", 10 * time.Second},
			{"10h11m12", (10 * time.Hour) +
				(11 * time.Minute) +
				(12 * time.Second)},
			{"1h2m3s", (1 * time.Hour) +
				(2 * time.Minute) +
				(3 * time.Second)},
			{"p2y12m27dt15h42m36", (2 * Year) +
				(12 * Month) +
				(27 * Day) +
				(15 * time.Hour) +
				(42 * time.Minute) +
				(36 * time.Second)},
		} {
			CoreValueTest(t, test.s, ParseDuration, test.x)
		}
		t.Run("Error", func(t *testing.T) {
			for _, s := range []string{
				"foo",
			} {
				CoreValueError(t, s, ParseDuration)
			}
		})
	})
	t.Run("FixedPoint", func(t *testing.T) {
		for _, test := range []struct {
			s string
			x float32
		}{
			{"1.23", 1.23},
		} {
			CoreValueTest(t, test.s, ParseFixedPoint, test.x)
		}
		t.Run("Error", func(t *testing.T) {
			for _, s := range []string{
				"foo",
			} {
				CoreValueError(t, s, ParseFixedPoint)
			}
		})
	})
	t.Run("Integer", func(t *testing.T) {
		for _, test := range []struct {
			s string
			x uint32
		}{
			{"123", 123},
		} {
			CoreValueTest(t, test.s, ParseInteger, test.x)
		}
		t.Run("Error", func(t *testing.T) {
			for _, s := range []string{
				"foo",
				"-1",
				"4294967296",
			} {
				CoreValueError(t, s, ParseInteger)
			}
		})
	})
	t.Run("IPAddress", func(t *testing.T) {
		for _, test := range []struct {
			s string
			x netip.Addr
		}{
			{"*", netip.IPv4Unspecified()},
			{"10", netip.MustParseAddr("10.0.0.0")},
			{"1.2", netip.MustParseAddr("1.2.0.0")},
			{"1.2.3", netip.MustParseAddr("1.2.3.0")},
			{"1.2.3.4", netip.MustParseAddr("1.2.3.4")},
			{"::1", netip.IPv6Loopback()},
			{"fe80::1", netip.MustParseAddr("fe80::01")},
		} {
			CoreValueTest(t, test.s, ParseIPAddress, test.x)
		}
		t.Run("Error", func(t *testing.T) {
			for _, s := range []string{
				"foo",
				"127.0.0.257",
			} {
				CoreValueError(t, s, ParseIPAddress)
			}
		})
	})
	t.Run("NetPrefix", func(t *testing.T) {
		for _, test := range []struct {
			s string
			x netip.Prefix
		}{
			{"*/0", netip.PrefixFrom(netip.IPv4Unspecified(), 0)},
			{"*/53", netip.PrefixFrom(netip.IPv4Unspecified(), 53)},
			{"::1/53", netip.PrefixFrom(netip.IPv6Loopback(), 53)},
			{"10/8", netip.MustParsePrefix("10.0.0.0/8")},
			{"1.2/16", netip.MustParsePrefix("1.2.0.0/16")},
			{"1.2.3/24", netip.MustParsePrefix("1.2.3.0/24")},
		} {
			CoreValueTest(t, test.s, ParseNetPrefix, test.x)
		}
		t.Run("Error", func(t *testing.T) {
			for _, s := range []string{
				"foo",
				"::1/65536",
			} {
				CoreValueError(t, s, ParseNetPrefix)
			}
		})
	})
	t.Run("Port", func(t *testing.T) {
		for _, test := range []struct {
			s string
			x uint16
		}{
			{"*", 0},
			{"123", 123},
		} {
			CoreValueTest(t, test.s, ParsePort, test.x)
		}
		t.Run("Error", func(t *testing.T) {
			for _, s := range []string{
				"foo",
				"-123",
				"65536",
			} {
				CoreValueError(t, s, ParsePort)
			}
		})
	})
	t.Run("Size", func(t *testing.T) {
		for _, test := range []struct {
			s string
			x uint64
		}{
			{"123", 123},
			{"123k", 123 * 1024},
			{"123m", 123 * 1024 * 1024},
			{"123g", 123 * 1024 * 1024 * 1024},
		} {
			CoreValueTest(t, test.s, ParseSize, test.x)
		}
		t.Run("Error", func(t *testing.T) {
			for _, s := range []string{
				"foo",
				"-123",
				"123x",
			} {
				CoreValueError(t, s, ParseSize)
			}
		})
	})
}
