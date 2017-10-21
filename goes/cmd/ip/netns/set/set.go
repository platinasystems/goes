// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package set

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/netns"
	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

const (
	Name    = "set"
	Apropos = "set/unset network namespace identifier"
	Usage   = `ip netns set NETNSNAME NETNSID`
	Man     = `
SEE ALSO
	ip man netns || ip netns -man
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
	var name string
	var nsid int32

	_, args = options.New(args)

	switch len(args) {
	case 0:
		return fmt.Errorf("NETNSNAME: missing")
	case 1:
		return fmt.Errorf("NETNSID: missing")
	case 2:
		_, err := fmt.Sscan(args[1], &nsid)
		if err != nil {
			return fmt.Errorf("NETNSID: %v", err)
		}
		name = args[0]
	default:
		return fmt.Errorf("%v: unexpected", args[2:])
	}

	f, err := os.Open(filepath.Join(rtnl.VarRunNetns, name))
	if err != nil {
		return err
	}
	defer f.Close()

	sock, err := nl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	req, err := nl.NewMessage(
		nl.Hdr{
			Type:  rtnl.RTM_NEWNSID,
			Flags: nl.NLM_F_REQUEST | nl.NLM_F_ACK,
		},
		rtnl.NetnsMsg{
			Family: rtnl.AF_UNSPEC,
		},
		nl.Attr{rtnl.NETNSA_FD, nl.Uint32Attr(f.Fd())},
		nl.Attr{rtnl.NETNSA_NSID, nl.Int32Attr(nsid)},
	)
	if err != nil {
		return err
	}

	return nl.NewSockReceiver(sock).UntilDone(req, func(b []byte) {})
}

func (Command) Complete(args ...string) (list []string) {
	if n := len(args); n == 1 {
		list = netns.CompleteName(args[n-1])
	}
	return
}
