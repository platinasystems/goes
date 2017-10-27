// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package add

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/genl"
	"github.com/platinasystems/go/internal/nl/genl/fou"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

const (
	Name    = "add"
	Apropos = "add a Foo-over-UDP receive port"
	Usage   = "ip foo add port PORT [ gue | ipproto PROTO ]"
	Man     = `
OPTIONS
	port PORT
		UDP listening port

	gue	use generic UDP encapsulation

	ipproto PROTO
		encapsulated protocol { 4 (ip, default) | 6 (ipv6) | 47 (gre) }

SEE ALSO
	ip fou man delete || ip fou delete -man
	ip man fou || ip fou -man
	man ip || ip -man`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	var attrs []nl.Attr

	opt, args := options.New(args)
	args = opt.Flags.More(args, "gue")
	args = opt.Parms.More(args, "port", "ipproto")
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	if s := opt.Parms.ByName["port"]; len(s) > 0 {
		var port uint16
		if _, err := fmt.Sscan(s, &port); err != nil {
			return fmt.Errorf("port %q %v", s, err)
		}
		attrs = append(attrs, nl.Attr{fou.FOU_ATTR_PORT,
			nl.Be16Attr(port)})
	} else {
		return fmt.Errorf("missing port")
	}

	t := fou.FOU_ENCAP_DIRECT
	if opt.Flags.ByName["gue"] {
		t = fou.FOU_ENCAP_GUE
	}
	attrs = append(attrs, nl.Attr{fou.FOU_ATTR_TYPE, nl.Uint8Attr(t)})

	if opt.Parms.ByName["-f"] == "inet6" {
		attrs = append(attrs, nl.Attr{fou.FOU_ATTR_AF,
			nl.Uint16Attr(rtnl.AF_INET6)})
	}

	if t == fou.FOU_ENCAP_DIRECT {
		var ipproto uint8
		switch s := opt.Parms.ByName["ipproto"]; s {
		case "", "4", "ip", "inet":
			ipproto = 4
		case "6", "ip6", "ipv6", "inet6":
			ipproto = 6
		case "47", "gre":
			ipproto = 47
		default:
			return fmt.Errorf("iproto %q unknown", s)
		}
		attrs = append(attrs, nl.Attr{fou.FOU_ATTR_IPPROTO,
			nl.Uint8Attr(ipproto)})
	}

	sock, err := nl.NewSock(nl.NETLINK_GENERIC)
	if err != nil {
		return err
	}
	defer sock.Close()

	sr := nl.NewSockReceiver(sock)

	nlt, err := genl.GetFamily(sr, fou.FOU_GENL_NAME)
	if err != nil {
		return err
	}
	req, err := nl.NewMessage(nl.Hdr{
		Type:  nlt,
		Flags: nl.NLM_F_REQUEST | nl.NLM_F_ACK,
	}, genl.Msg{
		Cmd:     fou.FOU_CMD_ADD,
		Version: fou.FOU_GENL_VERSION,
	}, attrs...)
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
