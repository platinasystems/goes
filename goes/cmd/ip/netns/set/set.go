// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package set

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "set"
	Apropos = "set/unset network namespace identifier"
	Usage   = `
	ip netns set NETNSNAME NETNSID
	`
	Man = `
SEE ALSO
	ip man netns || ip netns -man
`
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

	sock, err := rtnl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	req, err := rtnl.NewMessage(
		rtnl.Hdr{
			Type:  rtnl.RTM_NEWNSID,
			Flags: rtnl.NLM_F_REQUEST | rtnl.NLM_F_ACK,
		},
		rtnl.NetnsMsg{
			Family: rtnl.AF_UNSPEC,
		},
		rtnl.Attr{rtnl.NETNSA_FD, rtnl.Uint32Attr(f.Fd())},
		rtnl.Attr{rtnl.NETNSA_NSID, rtnl.Int32Attr(nsid)},
	)
	if err != nil {
		return err
	}

	return rtnl.NewSockReceiver(sock).UntilDone(req, func(b []byte) {})
}
