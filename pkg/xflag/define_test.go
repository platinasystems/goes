// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xflag

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/netip"
	"testing"
	"time"
)

type xint int

var xints = []string{"zero", "one", "two", "three", "four", "five"}

func (x xint) String() string {
	if 0 <= x && x < xint(len(xints)) {
		return xints[x]
	}
	return fmt.Sprint(int(x))
}

func (x *xint) UnmarshalText(data []byte) error {
	text := string(data)
	for i, s := range xints {
		if s == text {
			*x = xint(i)
			return nil
		}
	}
	return errors.New("unknown")
}

func defineResult(t *testing.T, success bool, args ...any) {
	t.Helper()
	if success {
		t.Log(args...)
	} else {
		t.Error(args...)
	}
}

func TestDefine(t *testing.T) {
	var (
		bf = false
		bt = true
		i  = 1
		u  = 5
		s  = "foobar"
		x  = xint(0)
		d  = 3 * time.Second
		a  = netip.IPv6Loopback()
		ap = netip.AddrPortFrom(netip.IPv6Loopback(), 8080)
		ha = net.HardwareAddr{0, 0, 0, 0, 0, 0}

		exp = struct {
			a   netip.Addr
			apa netip.Addr
			ha  net.HardwareAddr
		}{
			a:   netip.AddrFrom4([4]byte{192, 168, 1, 1}),
			apa: netip.MustParseAddr("fc00:1234::2"),
			ha:  net.HardwareAddr{0, 0x11, 0x22, 0x33, 0x44, 0x55},
		}
	)
	fs := flag.NewFlagSet("define-help", flag.ContinueOnError)
	DefineIn(fs, &bf, "f", "")
	DefineIn(fs, &bt, "t", "")
	DefineIn(fs, &i, "i", "")
	DefineIn(fs, &u, "u", "")
	DefineIn(fs, &s, "s", "")
	DefineIn(fs, &x, "x", "")
	DefineIn(fs, &d, "d", "")
	DefineIn(fs, &a, "a", "")
	DefineIn(fs, &ap, "ap", "")
	DefineIn(fs, &ha, "ha", "")
	err := fs.Parse([]string{
		"-f",
		"-t=false",
		"-i", "-3",
		"-u", "7",
		"-x", "two",
		"-s", "goo",
		"-d", "5ms",
		"-a", "192.168.1.1",
		"-ap", "[fc00:1234::2]:8080",
		"-ha", "00:11:22:33:44:55",
	})
	if err != nil {
		t.Fatal(err)
	}
	defineResult(t, bf, "f:", bf)
	defineResult(t, !bt, "t:", bt)
	defineResult(t, i == -3, "i:", i)
	defineResult(t, u == 7, "u:", u)
	defineResult(t, x == 2, "x:", x)
	defineResult(t, d == 5*time.Millisecond, "d:", d)
	defineResult(t, a.Compare(exp.a) == 0, "a:", a)
	defineResult(t, ap.Addr().Compare(exp.apa) == 0 && ap.Port() == 8080,
		"ap:", ap)
	defineResult(t, bytes.Compare(ha, exp.ha) == 0, "ha:", ha)
}
