// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package set

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/platinasystems/goes/cmd/ip/internal/options"
	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/goes/internal/netns"
	"github.com/platinasystems/goes/internal/nl"
	"github.com/platinasystems/goes/internal/nl/rtnl"
)

type Command struct{}

func (Command) String() string { return "set" }

func (Command) Usage() string {
	return `ip netns set NETNSNAME NETNSID`
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "set/unset network namespace identifier",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
SEE ALSO
	ip man netns || ip netns -man
	man ip || ip -man`,
	}
}

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

	return nl.NewSockReceiver(sock).UntilDone(req, nl.DoNothing)
}

func (Command) Complete(args ...string) (list []string) {
	if n := len(args); n == 1 {
		list = netns.CompleteName(args[n-1])
	}
	return
}
