// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package listid

import (
	"fmt"
	"io/ioutil"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "list-id"
	Apropos = "list network namespace identifiers"
	Usage   = `ip netns list-id NETNSNAME`
	Man     = `
SEE ALSO
	ip man netns || ip netns -man
	man ip || ip -man`
)

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}

var man = lang.Alt{
	lang.EnUS: Man,
}

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	_, args = options.New(args)
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}
	namebyid := make(map[int32]string)
	sock, err := rtnl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	sr := rtnl.NewSockReceiver(sock)

	varRunNetns, err := ioutil.ReadDir(rtnl.VarRunNetns)
	if err == nil {
		for _, fi := range varRunNetns {
			nsid, err := sr.Nsid(fi.Name())
			if err == nil && nsid >= 0 {
				namebyid[nsid] = fi.Name()
			}
		}
	}
	req, err := rtnl.NewMessage(
		rtnl.Hdr{
			Type:  rtnl.RTM_GETNSID,
			Flags: rtnl.NLM_F_REQUEST | rtnl.NLM_F_DUMP,
		},
		rtnl.RtGenMsg{
			Family: rtnl.AF_UNSPEC,
		},
	)
	if err != nil {
		return err
	}
	return sr.UntilDone(req, func(b []byte) {
		var netnsa rtnl.Netnsa
		t := rtnl.HdrPtr(b).Type
		if t != rtnl.RTM_DELNSID && t != rtnl.RTM_NEWNSID {
			return
		}
		if n, err := netnsa.Write(b); err != nil || n == 0 {
			return
		}
		if t == rtnl.RTM_DELNSID {
			fmt.Print("Deleted ")
		}
		nsid := int32(-1)
		if val := netnsa[rtnl.NETNSA_NSID]; len(val) > 0 {
			nsid = rtnl.Int32(val)
		}
		fmt.Print(nsid)
		if name, found := namebyid[nsid]; found {
			fmt.Print(": ", name)
		}
		fmt.Println()
	})
}
