// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package delete

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
	Name    = "delete"
	Apropos = "delete a Foo-over-UDP receive port"
	Usage   = "ip foo del port PORT"
	Man     = `
OPTIONS
	port PORT
		UDP listening port

SEE ALSO
	ip fou man add || ip fou add -man
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

	if opt.Parms.ByName["-f"] == "inet6" {
		attrs = append(attrs, nl.Attr{fou.FOU_ATTR_AF,
			nl.Uint16Attr(rtnl.AF_INET6)})
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
		Cmd:     fou.FOU_CMD_DEL,
		Version: fou.FOU_GENL_VERSION,
	}, attrs...)
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
