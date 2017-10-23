// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package delete

import (
	"fmt"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

const (
	Name    = "delete"
	Apropos = "remove virtual link"
	Usage   = "ip link delete DEVICE"
)

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	var (
		hdr nl.Hdr
		msg rtnl.IfInfoMsg
	)
	if n := len(args); n == 0 {
		return fmt.Errorf("missing DEVICE")
	} else if n > 1 {
		return fmt.Errorf("%v unexpected", args[1:])
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

	index, found := rtnl.If.IndexByName[args[0]]
	if !found {
		return fmt.Errorf("%q not found", args[0])
	}
	hdr.Flags = nl.NLM_F_REQUEST | nl.NLM_F_ACK
	hdr.Type = rtnl.RTM_DELLINK
	msg.Family = rtnl.AF_UNSPEC
	msg.Index = index
	req, err := nl.NewMessage(hdr, msg)
	if err == nil {
		err = sr.UntilDone(req, func([]byte) {})
	}
	return err
}
