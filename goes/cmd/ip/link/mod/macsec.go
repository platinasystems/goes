// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

// ip link COMMAND type macsec
//	{ [ address LLADDR ] port { 1..2^16-1 } | sci <u64> }
//	[ cipher { default | gcm-aes-128 } ]
//	[ icvlen { 8..16 } ]
//	[ [no-]encrypt ]
//	[ [no-]send-sci ]
//	[ [no-]end-station ]
//	[ [no-]scb ]
//	[ [no-]protect ]
//	[ [no-]replay window { 0..2^32-1 } ]
//	[ validate { strict | check | disabled } ]
//	[ encodingsa { 0..3 } ]
func (m *mod) parseTypeMacSec() error {
	var s string
	var err error

	m.args = m.opt.Parms.More(m.args,
		"address", // LLADDR
		"port",    // PORT
		"sci",     // SCI
	)
	if s = m.opt.Parms.ByName["port"]; len(s) > 0 {
		var port uint16
		if _, err = fmt.Sscan(s, &port); err != nil {
			return fmt.Errorf("port: %q %v", s, err)
		}
		m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_MACSEC_PORT,
			rtnl.Be16Attr(port)})
	} else if s = m.opt.Parms.ByName["sci"]; len(s) > 0 {
		var sci uint64
		if _, err = fmt.Sscan(s, &sci); err != nil {
			return fmt.Errorf("sci: %q %v", s, err)
		}
		m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_MACSEC_SCI,
			rtnl.Be64Attr(sci)})
	} else {
		return fmt.Errorf("missing port or sci")
	}

	for _, x := range []struct {
		set   []string
		unset []string
		t     uint16
	}{
		{
			[]string{"encrypt", "+encrypt"},
			[]string{"no-encrypt", "-encrypt"},
			rtnl.IFLA_MACSEC_ENCRYPT,
		},
		{
			[]string{"send-sci", "+send-sci"},
			[]string{"no-send-sci", "-send-sci"},
			rtnl.IFLA_MACSEC_INC_SCI,
		},
		{
			[]string{"end-station", "+end-station"},
			[]string{"no-end-station", "-end-station"},
			rtnl.IFLA_MACSEC_ES,
		},
		{
			[]string{"scb", "+scb"},
			[]string{"no-scb", "-scb"},
			rtnl.IFLA_MACSEC_SCB,
		},
		{
			[]string{"protect", "+protect"},
			[]string{"no-protect", "-protect"},
			rtnl.IFLA_MACSEC_PROTECT,
		},
		{
			[]string{"replay", "+replay"},
			[]string{"no-replay", "-replay"},
			rtnl.IFLA_MACSEC_REPLAY_PROTECT},
	} {
		m.args = m.opt.Flags.More(m.args, x.set, x.unset)
		if m.opt.Flags.ByName[x.set[0]] {
			m.tinfo = append(m.tinfo, rtnl.Attr{x.t,
				rtnl.Uint8Attr(1)})
		} else if m.opt.Flags.ByName[x.unset[0]] {
			m.tinfo = append(m.tinfo, rtnl.Attr{x.t,
				rtnl.Uint8Attr(0)})
		}
	}
	if m.opt.Flags.ByName["replay"] {
		var window uint32
		m.args = m.opt.Parms.More(m.args, "window")
		s = m.opt.Parms.ByName["window"]
		if len(s) == 0 {
			return fmt.Errorf("missing window")
		}
		if _, err := fmt.Sscan(s, &window); err != nil {
			return fmt.Errorf("window: %q %v", s, err)
		}
		m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_MACSEC_WINDOW,
			rtnl.Uint32Attr(window)})
	}
	m.args = m.opt.Parms.More(m.args,
		"cipher",   // SUITE
		"validate", // { strict | check | disabled }
	)
	if s = m.opt.Parms.ByName["cipher"]; len(s) > 0 {
		switch s {
		case "default", "gcm-aes-128":
			id := rtnl.MACSEC_DEFAULT_CIPHER_ID
			m.tinfo = append(m.tinfo,
				rtnl.Attr{rtnl.IFLA_MACSEC_CIPHER_SUITE,
					rtnl.Uint64Attr(id)})
		default:
			return fmt.Errorf("cipher: %q unknown", s)
		}
	}
	if s = m.opt.Parms.ByName["validate"]; len(s) > 0 {
		validate, found := map[string]uint8{
			"disabled": rtnl.MACSEC_VALIDATE_DISABLED,
			"check":    rtnl.MACSEC_VALIDATE_CHECK,
			"strict":   rtnl.MACSEC_VALIDATE_STRICT,
		}[s]
		if !found {
			return fmt.Errorf("validate: %q unkown", s)
		}
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_MACSEC_VALIDATION,
				rtnl.Uint8Attr(validate)})
	}
	for _, x := range []struct {
		names []string
		t     uint16
	}{
		{[]string{"icvlen", "icv-len"}, rtnl.IFLA_MACSEC_ICV_LEN},
		{[]string{"encodingsa", "encoding-sa"},
			rtnl.IFLA_MACSEC_ENCODING_SA},
	} {
		var u8 uint8
		m.args = m.opt.Parms.More(m.args, x.names)
		s := m.opt.Parms.ByName[x.names[0]]
		if len(s) == 0 {
			continue
		}
		if _, err = fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("%s: %q %v", x.names[0], s, err)
		}
		m.tinfo = append(m.tinfo, rtnl.Attr{x.t, rtnl.Uint8Attr(u8)})
	}
	return nil
}
