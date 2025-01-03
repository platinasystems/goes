// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_conf

import (
	"net/netip"
	"testing"
)

func TestAddressMatchList(t *testing.T) {
	input := Block{
		Statement{"1.2.3.4"},
		Statement{"5.6.7.8"},
		Statement{"46/24"},
		Block{
			Statement{"!", "1.2.3.4"},
			Statement{"1.2.3/24"},
		},
		Statement{"10/8"},
		Statement{"!", "1.2.3/24"},
		Block{
			Statement{"1.2/16"},
			Statement{"3/8"},
		},
		Block{
			Statement{"any"},
			Statement{"key", "foo"},
		},
	}
	expect := AddressMatchList{
		AddressMatchElement{
			Match: netip.MustParseAddr("1.2.3.4"),
		},
		AddressMatchElement{
			Match: netip.MustParseAddr("5.6.7.8"),
		},
		AddressMatchElement{
			Match: netip.MustParsePrefix("46.0.0.0/24"),
		},
		AddressMatchElement{
			Match: AddressMatchList{
				AddressMatchElement{
					Not:   true,
					Match: netip.MustParseAddr("1.2.3.4"),
				},
				AddressMatchElement{
					Match: netip.MustParsePrefix("1.2.3.0/24"),
				},
			},
		},
		AddressMatchElement{
			Match: netip.MustParsePrefix("10.0.0.0/8"),
		},
		AddressMatchElement{
			Not:   true,
			Match: netip.MustParsePrefix("1.2.3.0/24"),
		},
		AddressMatchElement{
			Match: AddressMatchList{
				AddressMatchElement{
					Match: netip.MustParsePrefix("1.2.0.0/16"),
				},
				AddressMatchElement{
					Match: netip.MustParsePrefix("3.0.0.0/8"),
				},
			},
		},
		AddressMatchElement{
			Match: AddressMatchList{
				AddressMatchElement{
					Match: "any",
				},
				AddressMatchElement{
					Match: AddressMatchKey("foo"),
				},
			},
		},
	}
	tlog = t.Log
	tlogf = t.Logf
	aml, err :=
		input.AddressMatchList()
	if err != nil {
		t.Error(err)
	} else if err = aml.verify(expect); err != nil {
		t.Error(err)
	} else {
		t.Log(aml)
	}
}
