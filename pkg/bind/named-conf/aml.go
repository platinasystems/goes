// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_conf

import (
	"fmt"
	"net/netip"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

// An AddressMatchElement may be defined by,
//
//	[ ! ] ( <ip_address> | <netprefix> | key <server_key> | <acl_name> |
//		{ address_match_list } )
//
// Where,
//   - ip_address: an IP address (IPv4 or IPv6)
//   - netprefix: an IP prefix (in / notation)
//   - server_key: a key ID, as defined by the key statement
//   - acl_name: the name of an address match list defined by an acl statement
//
// e.g.
//
//	options {
//		forwarders {
//			1.2.3.4;
//			5.6.7.8;
//		};
//		keep-response-order { 46/24; };
//		listen-on port 1234 {
//			!1.2.3.4;
//			1.2.3 / 24;
//		};
//		topology {
//			10/8;
//			!1.2.3/24;
//			{
//				1.2/16;
//				3/8;
//			};
//		};
//	};
//	controls {
//		inet 10.0.0.1 allow {
//			any;
//			key foo;
//		};
//	}
type AddressMatchElement struct {
	Not bool
	// [netip.Addr], [netip.Prefix], [AddressMatchKey], string (acl),
	// or  [AddressMatchList].
	Match any
}

func (stmt Statement) AddressMatchElement() (
	ame AddressMatchElement, err error,
) {
	if len(stmt) == 0 {
		return
	}
	if block, ok := stmt[0].(Block); ok {
		ame.Match, err = block.AddressMatchList()
		return
	}
	s, ok := stmt[0].(string)
	if !ok {
		err = xerrors.Invalid("address match element")
		return
	}
	if s == "!" {
		ame.Not = true
		if stmt = stmt[1:]; len(stmt) == 0 {
			err = xerrors.Incomplete("address match element")
			return
		} else if s, ok = stmt[0].(string); !ok {
			err = ErrSyntax
			return
		}
	}
	if s == "key" {
		if len(stmt) == 1 {
			err = xerrors.Incomplete("address match key")
		} else if s, ok := stmt[1].(string); !ok {
			err = xerrors.Invalid("address match key")
		} else {
			ame.Match = AddressMatchKey(s)
		}
	} else if strings.Contains(s, "/") {
		ame.Match, err = ParseNetPrefix(s)
	} else if r, n := utf8.DecodeRuneInString(s); n == 0 {
		err = xerrors.Invalid("address match element")
	} else if unicode.IsDigit(r) || strings.ContainsAny(s, ":.") {
		ame.Match, err = ParseIPAddress(s)
	} else {
		ame.Match = s
	}
	return
}

func (ame AddressMatchElement) Format(w fmt.State, verb rune) {
	if ame.Not {
		fmt.Fprint(w, "!")
	}
	fmt.Fprint(w, ame.Match)
}

func (ame AddressMatchElement) verify(expect AddressMatchElement) error {
	mismatch := func() error {
		return fmt.Errorf("MISMATCH\nhave: %v\nwant: %v",
			ame.Match, expect.Match)
	}
	if (ame.Not && !expect.Not) || (!ame.Not && expect.Not) {
		return fmt.Errorf("MISMATCHED NEGATION: %v", ame)
	}
	switch t := ame.Match.(type) {
	case AddressMatchList:
		if l, ok := expect.Match.(AddressMatchList); !ok {
			return mismatch()
		} else {
			return t.verify(l)
		}
	case AddressMatchKey:
		if k, ok := expect.Match.(AddressMatchKey); !ok || t != k {
			return mismatch()
		}
	case netip.Addr:
		if a, ok := expect.Match.(netip.Addr); !ok || t != a {
			return mismatch()
		}
	case netip.Prefix:
		if p, ok := expect.Match.(netip.Prefix); !ok || t != p {
			return mismatch()
		}
	case string:
		if s, ok := expect.Match.(string); !ok || t != s {
			return mismatch()
		}
	default:
		return mismatch()
	}
	return nil
}

type AddressMatchKey string

func (k AddressMatchKey) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "key ", string(k))
}

type AddressMatchList []AddressMatchElement

func (block Block) AddressMatchList() (AddressMatchList, error) {
	var aml AddressMatchList
	for _, v := range block {
		switch t := v.(type) {
		case Block:
			if sub, err := t.AddressMatchList(); err != nil {
				return aml, err
			} else if sub != nil {
				aml = append(aml, AddressMatchElement{
					Match: sub,
				})
			}
		case Statement:
			if ame, err := t.AddressMatchElement(); err != nil {
				return aml, err
			} else if ame.Match != nil {
				aml = append(aml, ame)
			}
		default:
			return aml, ErrSyntax
		}
	}
	return aml, nil
}

func (aml AddressMatchList) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "{")
	iw := NewIndent(w)
	for _, ame := range aml {
		fmt.Fprint(iw, "\n", ame, ";")
	}
	fmt.Fprint(w, "\n}")
}

func (aml AddressMatchList) verify(expect AddressMatchList) error {
	for i, ame := range aml {
		if i >= len(expect) {
			return fmt.Errorf("UNEXPECTED: %v", ame)
		}
		if err := ame.verify(expect[i]); err != nil {
			return err
		}
	}
	if len(aml) < len(expect) {
		return fmt.Errorf("MISSING: %v", expect[len(aml)])
	}
	return nil
}
