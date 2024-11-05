// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsdb

import (
	"bufio"
	"context"
	"fmt"
	"maps"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
)

var (
	ErrIncomplete  = xerrors.ErrIncomplete
	ErrUnsupported = xerrors.ErrUnsupported
)

type UniqueString = xdnsmessage.UniqueString

var MakeUniqueString = xdnsmessage.MakeUniqueString
var Root = MakeUniqueString(".")

// This only retains INET Resource Records;
// it ignores CSNET, CHAOS, and HESIOD.
type DB interface {
	// Scan Resource Records until EOF or cancelled.
	Fread(context.Context, *os.File) error
	// Lookup Resource Records of the named host within the given zone.
	// If zone is empty and name doesn't have a root (".") suffix,
	// return records of the named host within the default zone;
	// otherwise, return records of the first matching zone suffix.
	Lookup(zone, name string) []RR
	// Calls given function foreach (zone, name) in domain order.
	Range(func(zone, name string, rrs []RR) bool)
	WhoIsAt(netip.Addr) (zone, name string)
}

type RR struct {
	TTL time.Time
	xdnsmessage.Resource
}

func MakeDB(zone string) DB {
	if len(zone) == 0 {
		zone = "."
	} else if !strings.HasSuffix(zone, ".") {
		zone += "."
	}
	db := new(db)
	db.zone = MakeUniqueString(strings.ToLower(zone))
	db.zones = make(map[UniqueString]map[UniqueString][]RR)
	db.zones[db.zone] = make(map[UniqueString][]RR)
	if db.zone != Root {
		db.zones[Root] = make(map[UniqueString][]RR)
	}
	db.zns = make(map[netip.Addr]zn)
	return db
}

type db struct {
	sync.Mutex
	zones map[UniqueString]map[UniqueString][]RR
	zns   map[netip.Addr]zn
	zone  UniqueString
}

type zn struct{ zone, name UniqueString }

func (db *db) Fread(ctx context.Context, f *os.File) error {
	db.Lock()
	defer db.Unlock()

	return db.fread(ctx, f, db.zone)
}

func (db *db) fread(
	ctx context.Context, f *os.File, origin UniqueString,
) error {
	var defaultTTL time.Duration
	var tokens []string
	var cont, uselast bool

	fn := filepath.Base(f.Name())
	now := time.Now()
	last := xdnsmessage.UniqueEmptyString
	sc := bufio.NewScanner(f)
	firstln := true
	for lno := 1; ctx.Err() == nil && sc.Scan(); lno++ {
		s := sc.Text()
		ts := strings.TrimSpace(s)
		if len(s) == 0 || strings.HasPrefix(ts, ";") {
			continue
		}
		if firstln {
			firstln = false
			uselast = strings.HasPrefix(s, " ") ||
				strings.HasPrefix(s, "\t")
		}
		for _, s = range strings.Fields(ts) {
			if strings.HasPrefix(s, ";") {
				break
			}
			if !cont {
				if strings.HasPrefix(s, "(") {
					s = strings.TrimPrefix(s, "(")
					cont = true
				}
				if len(s) > 0 {
					tokens = append(tokens, s)
				}
			} else {
				if strings.HasSuffix(s, ")") {
					s = strings.TrimSuffix(s, ")")
					cont = false
				}
				if len(s) > 0 {
					tokens = append(tokens, s)
				}
			}
		}
		if cont || len(tokens) == 0 {
			continue
		}
		firstln = true
		if strings.HasPrefix(tokens[0], "$") {
			s := strings.ToUpper(strings.TrimPrefix(tokens[0], "$"))
			switch s {
			case "INCLUDE":
				if len(tokens) < 2 {
					return fmt.Errorf("%s[%d]:INCLUDE: %w",
						fn, lno, ErrIncomplete)
				}
				include, err := os.Open(tokens[1])
				if err != nil {
					return fmt.Errorf("%s[%d]: %w",
						fn, lno, err)
				}
				zone := origin
				if len(tokens) > 2 {
					s := strings.ToLower(tokens[2])
					zone = MakeUniqueString(s)
				}
				err = db.fread(ctx, include, zone)
				include.Close()
				if err != nil {
					return err
				}
			case "ORIGIN":
				if len(tokens) == 1 || len(tokens[1]) == 0 {
					return fmt.Errorf("%s[%d]:ORIGIN: %w",
						fn, lno, ErrIncomplete)

				}
				s := strings.ToLower(tokens[1])
				origin = MakeUniqueString(s)
			case "TTL":
				if len(tokens) == 1 {
					return fmt.Errorf("%s[%d]:TTL: %w",
						fn, lno, ErrIncomplete)
				}
				u, err := strconv.ParseUint(tokens[1], 10, 32)
				if err != nil {
					return fmt.Errorf("%s[%d]:TTL: %w",
						fn, lno, err)
				}
				defaultTTL = time.Second * time.Duration(u)
			}
			tokens = tokens[:0]
			continue
		}
		var zone, name UniqueString
		var ttl time.Duration
		c := xdnsmessage.Class0
		if uselast {
			uselast = false
			if s := last.String(); strings.HasSuffix(s, ".") {
				zone = Root
				s = strings.TrimSuffix(s, ".")
				name = MakeUniqueString(strings.ToLower(s))
			} else {
				zone = origin
				name = last
			}
		} else if tokens[0] == "." {
			zone = Root
			name = MakeUniqueString("@")
			tokens = tokens[1:]
		} else if strings.HasSuffix(tokens[0], ".") {
			zone = Root
			s := strings.TrimSuffix(tokens[0], ".")
			name = MakeUniqueString(strings.ToLower(s))
			last = name
			tokens = tokens[1:]
		} else {
			zone = origin
			name = MakeUniqueString(strings.ToLower(tokens[0]))
			last = name
			tokens = tokens[1:]
		}
		for len(tokens) > 0 {
			if ttl == 0 {
				u, err := strconv.ParseUint(tokens[0], 10, 32)
				if err == nil {
					ttl = time.Second * time.Duration(u)
					tokens = tokens[1:]
					continue
				}
			}
			if c == xdnsmessage.Class0 {
				tc, err := xdnsmessage.ClassNamed(tokens[0])
				if err == nil {
					c = tc
					tokens = tokens[1:]
					continue
				}
			}
			break
		}
		if ttl == 0 {
			ttl = defaultTTL
		}
		if c != xdnsmessage.Class0 && c != xdnsmessage.ClassINET {
			continue
		}
		switch len(tokens) {
		case 0:
			return fmt.Errorf("%s[%d]:TYPE: %w",
				fn, lno, ErrIncomplete)
		case 1:
			return fmt.Errorf("%s[%d]:resource: %w",
				fn, lno, ErrIncomplete)
		}
		t, err := xdnsmessage.TypeNamed(tokens[0])
		if err != nil {
			return fmt.Errorf("%s[%d]:TYPE: %w", fn, lno, err)
		}
		tokens = tokens[1:]
		rr := RR{TTL: now.Add(ttl)}
		rr.Resource, tokens, err = t.ParseResource(tokens)
		tokens = tokens[:0]
		m, ok := db.zones[zone]
		if !ok {
			m = make(map[UniqueString][]RR)
			db.zones[zone] = m
		}
		m[name] = append(m[name], rr)
		switch rr.Type() {
		case xdnsmessage.TypeA:
			addr := rr.Resource.(xdnsmessage.A).Addr
			db.zns[addr] = zn{zone, name}
		case xdnsmessage.TypeAAAA:
			addr := rr.Resource.(xdnsmessage.AAAA).Addr
			db.zns[addr] = zn{zone, name}
		}
	}
	return nil
}

// Lookup Resource Records of the named host within the given zone.
// If zone is empty and name doesn't have a root (".") suffix,
// return records of the named host within the default zone;
// otherwise, return records of the first matching zone suffix.
func (db *db) Lookup(zone, name string) []RR {
	db.Lock()
	defer db.Unlock()

	if len(zone) > 0 {
		return db.zones[MakeUniqueString(zone)][MakeUniqueString(name)]
	}
	if !strings.HasSuffix(name, ".") {
		return db.zones[db.zone][MakeUniqueString(name)]
	}
	for uz, names := range db.zones {
		if suf := uz.String(); strings.HasSuffix(name, suf) {
			s := strings.TrimSuffix(name, suf)
			s = strings.TrimSuffix(s, ".")
			if rrs, ok := names[MakeUniqueString(s)]; ok {
				return rrs
			}
		}
	}
	return nil
}

func (db *db) Range(f func(zone, name string, rrs []RR) bool) {
	db.Lock()
	defer db.Unlock()

	zones := slices.SortedFunc(maps.Keys(db.zones), db.cmp)
	for _, uz := range zones {
		names := slices.SortedFunc(maps.Keys(db.zones[uz]), db.cmp)
		for _, un := range names {
			if !f(uz.String(), un.String(), db.zones[uz][un]) {
				return
			}
		}
	}
}

func (db *db) WhoIsAt(addr netip.Addr) (zone, name string) {
	db.Lock()
	defer db.Unlock()

	if zn, ok := db.zns[addr]; ok {
		zone = zn.zone.String()
		name = zn.name.String()
	}
	return
}

// Compare domain names in right to left order. e.g.
//
//	foo.com < bar.edu
func (*db) cmp(us1, us2 UniqueString) int {
	dom1 := strings.Split(us1.String(), ".")
	dom2 := strings.Split(us2.String(), ".")
	i1 := len(dom1) - 1
	i2 := len(dom2) - 1
	if i1 >= 0 && len(dom1[i1]) == 0 {
		i1 -= 1
	}
	if i2 >= 0 && len(dom2[i2]) == 0 {
		i2 -= 1
	}
	for ; i1 >= 0 && i2 >= 0; i1, i2 = i1-1, i2-1 {
		if dom1[i1] != dom2[i2] {
			l1 := strings.ToLower(dom1[i1])
			l2 := strings.ToLower(dom2[i2])
			return strings.Compare(l1, l2)
		}
	}
	return len(dom1) - len(dom2)
}
