// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package macsec

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/request"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

type Command struct{}

func (Command) String() string { return "macsec" }

func (Command) Usage() string {
	return `
ip link add type macsec [ OPTIONS ]...`

}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "add an 802.1AE MAC-level encryption link",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
OPTIONS
	{ [ address LLADDR ] port { 1..2^16-1 } | sci <u64> }

	cipher { default | gcm-aes-128 }

	icvlen { 8..16 }
		length of the Integrity Check Value (ICV).

	[no-]encrypt
		authenticated encryption, or authenticity only

	[no-]send-sci
		include SCI in every packet, or only when it is necessary

	[no-]end-station
		End Station bit

	[no-]scb
		Single Copy Broadcast bit

	[no-]protect
		device MACsec protection

	[no-]replay window { 0..2^32-1 }
		device replay protection

	window SIZE
		replay window size

	validate { strict | check | disabled }

	encodingsa { 0..3 }
		active secure association for transmission

SEE ALSO
	ip link add type man TYPE || ip link add type TYPE -man
	ip link man add || ip link add -man
	man ip || ip -man`,
	}
}

func (Command) Main(args ...string) error {
	var info nl.Attrs
	var s string

	opt, args := options.New(args)
	args = opt.Flags.More(args,
		[]string{"encrypt", "+encrypt"},
		[]string{"no-encrypt", "-encrypt"},
		[]string{"send-sci", "+send-sci"},
		[]string{"no-send-sci", "-send-sci"},
		[]string{"end-station", "+end-station"},
		[]string{"no-end-station", "-end-station"},
		[]string{"scb", "+scb"},
		[]string{"no-scb", "-scb"},
		[]string{"protect", "+protect"},
		[]string{"no-protect", "-protect"},
		[]string{"replay", "+replay"},
		[]string{"no-replay", "-replay"},
	)
	args = opt.Parms.More(args,
		"address", // LLADDR
		"port",    // PORT
		"sci",     // SCI
		"window",
		"cipher",   // SUITE
		"validate", // { strict | check | disabled }
		[]string{"icvlen", "icv-len"},
		[]string{"encodingsa", "encoding-sa"},
	)
	err := opt.OnlyName(args)
	if err != nil {
		return err
	}

	sock, err := nl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	sr := nl.NewSockReceiver(sock)

	if err = rtnl.MakeIfMaps(sr); err != nil {
		return err
	}

	add, err := request.New(opt)
	if err != nil {
		return err
	}

	if s = opt.Parms.ByName["port"]; len(s) > 0 {
		var port uint16
		if _, err = fmt.Sscan(s, &port); err != nil {
			return fmt.Errorf("port: %q %v", s, err)
		}
		info = append(info, nl.Attr{rtnl.IFLA_MACSEC_PORT,
			nl.Be16Attr(port)})
	} else if s = opt.Parms.ByName["sci"]; len(s) > 0 {
		var sci uint64
		if _, err = fmt.Sscan(s, &sci); err != nil {
			return fmt.Errorf("sci: %q %v", s, err)
		}
		info = append(info, nl.Attr{rtnl.IFLA_MACSEC_SCI,
			nl.Be64Attr(sci)})
	} else {
		return fmt.Errorf("missing port or sci")
	}

	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"encrypt", rtnl.IFLA_MACSEC_ENCRYPT},
		{"send-sci", rtnl.IFLA_MACSEC_INC_SCI},
		{"end-station", rtnl.IFLA_MACSEC_ES},
		{"scb", rtnl.IFLA_MACSEC_SCB},
		{"protect", rtnl.IFLA_MACSEC_PROTECT},
		{"replay", rtnl.IFLA_MACSEC_REPLAY_PROTECT},
	} {
		if opt.Flags.ByName[x.name] {
			info = append(info, nl.Attr{x.t,
				nl.Uint8Attr(1)})
		} else if opt.Flags.ByName["no-"+x.name] {
			info = append(info, nl.Attr{x.t,
				nl.Uint8Attr(0)})
		}
	}
	if opt.Flags.ByName["replay"] {
		var window uint32
		s = opt.Parms.ByName["window"]
		if len(s) == 0 {
			return fmt.Errorf("missing window")
		}
		if _, err := fmt.Sscan(s, &window); err != nil {
			return fmt.Errorf("window: %q %v", s, err)
		}
		info = append(info, nl.Attr{rtnl.IFLA_MACSEC_WINDOW,
			nl.Uint32Attr(window)})
	}
	if s = opt.Parms.ByName["cipher"]; len(s) > 0 {
		switch s {
		case "default", "gcm-aes-128":
			id := rtnl.MACSEC_DEFAULT_CIPHER_ID
			info = append(info,
				nl.Attr{rtnl.IFLA_MACSEC_CIPHER_SUITE,
					nl.Uint64Attr(id)})
		default:
			return fmt.Errorf("cipher: %q unknown", s)
		}
	}
	if s = opt.Parms.ByName["validate"]; len(s) > 0 {
		validate, found := map[string]uint8{
			"disabled": rtnl.MACSEC_VALIDATE_DISABLED,
			"check":    rtnl.MACSEC_VALIDATE_CHECK,
			"strict":   rtnl.MACSEC_VALIDATE_STRICT,
		}[s]
		if !found {
			return fmt.Errorf("validate: %q unkown", s)
		}
		info = append(info, nl.Attr{rtnl.IFLA_MACSEC_VALIDATION,
			nl.Uint8Attr(validate)})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"icvlen", rtnl.IFLA_MACSEC_ICV_LEN},
		{"encodingsa", rtnl.IFLA_MACSEC_ENCODING_SA},
	} {
		var u8 uint8
		s := opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		if _, err = fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("%s: %q %v", x.name, s, err)
		}
		info = append(info, nl.Attr{x.t, nl.Uint8Attr(u8)})
	}

	add.Attrs = append(add.Attrs, nl.Attr{rtnl.IFLA_LINKINFO, nl.Attrs{
		nl.Attr{rtnl.IFLA_INFO_KIND, nl.KstringAttr("macsec")},
		nl.Attr{rtnl.IFLA_INFO_DATA, info},
	}})
	req, err := add.Message()
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
