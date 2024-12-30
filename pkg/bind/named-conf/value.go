// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_conf

import (
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

const (
	Year  = 365 * 24 * time.Hour
	Month = 30 * 24 * time.Hour
	Week  = 7 * 24 * time.Hour
	Day   = 24 * time.Hour
)

func ParseBool(s string) (bool, error) {
	switch s {
	case "0", "no", "false":
		return false, nil
	case "1", "yes", "true":
		return true, nil
	}
	return false, xerrors.Invalid("bool", s)
}

func ParseDuration(s string) (ttl time.Duration, err error) {
	const (
		plain       = "wWdDhHmMsS"
		iso8601date = "yYmMdDtT"
		iso8601time = "hHMmsS"
	)
	suffixes := plain
	if s[0] == 'p' || s[0] == 'P' {
		suffixes = iso8601date
		s = s[1:]
	}
	for len(s) > 0 {
		var n uint64
		var suffix rune
		if i := strings.IndexAny(s, suffixes); i < 0 {
			suffix = 's'
			n, err = strconv.ParseUint(s, 10, 64)
			s = s[:0]
		} else {
			suffix, _ = utf8.DecodeRuneInString(s[i:])
			if i > 0 {
				n, err = strconv.ParseUint(s[:i], 10, 64)
			}
			s = s[i+1:]
		}
		if err != nil {
			break
		}
		switch suffixes {
		case plain:
			switch suffix {
			case 'w', 'W':
				ttl += time.Duration(n) * Week
			case 'd', 'D':
				ttl += time.Duration(n) * Day
			case 'h', 'H':
				ttl += time.Duration(n) * time.Hour
			case 'm', 'M':
				ttl += time.Duration(n) * time.Minute
			case 's', 'S':
				ttl += time.Duration(n) * time.Second
			}
		case iso8601date:
			switch suffix {
			case 'y', 'Y':
				ttl += time.Duration(n) * Year
			case 'm', 'M':
				ttl += time.Duration(n) * Month
			case 'd', 'D':
				ttl += time.Duration(n) * Day
			case 't', 'T':
				suffixes = iso8601time
			}
		case iso8601time:
			switch suffix {
			case 'h', 'H':
				ttl += time.Duration(n) * time.Hour
			case 'm', 'M':
				ttl += time.Duration(n) * time.Minute
			case 's', 'S':
				ttl += time.Duration(n) * time.Second
			}
		}
	}
	return
}

func ParseFixedPoint(s string) (float32, error) {
	f64, err := strconv.ParseFloat(s, 32)
	return float32(f64), err
}

func ParseInteger(s string) (uint32, error) {
	u64, err := strconv.ParseUint(s, 10, 32)
	return uint32(u64), err
}

func ParseIPAddress(s string) (netip.Addr, error) {
	if s == "*" {
		return netip.IPv4Unspecified(), nil
	}
	if strings.Index(s, ":") >= 0 {
		return netip.ParseAddr(s)
	}
	switch len(strings.Split(s, ".")) {
	case 1:
		s += ".0.0.0"
	case 2:
		s += ".0.0"
	case 3:
		s += ".0"
	}
	return netip.ParseAddr(s)
}

func ParseNetPrefix(s string) (netip.Prefix, error) {
	slash := strings.Index(s, "/")
	if slash < 0 {
		return netip.Prefix{}, fmt.Errorf(`%q: missing "/BITS`, s)
	}
	bits, err := strconv.ParseUint(s[slash+1:], 10, 8)
	if err != nil {
		return netip.Prefix{}, err
	}
	addr, err := ParseIPAddress(s[:slash])
	if err != nil {
		return netip.Prefix{}, err
	}
	return netip.PrefixFrom(addr, int(bits)), nil
}

func ParsePort(s string) (uint16, error) {
	if s == "*" {
		return 0, nil
	}
	u64, err := strconv.ParseUint(s, 10, 16)
	return uint16(u64), err
}

func ParseSize(s string) (uint64, error) {
	var suffix rune
	if i := strings.IndexAny(s, "kKmMgG"); i > 0 {
		suffix, _ = utf8.DecodeRuneInString(s[i:])
		s = s[:i]
	}
	sz, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return sz, err
	}
	switch suffix {
	case 'k', 'K':
		sz *= 1024
	case 'm', 'M':
		sz *= 1024 * 1024
	case 'g', 'G':
		sz *= 1024 * 1024 * 1024
	}
	return sz, nil
}
