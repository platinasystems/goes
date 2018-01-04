// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package listid

import (
	"fmt"
	"io/ioutil"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

type Command struct{}

func (Command) String() string { return "list-id" }

func (Command) Usage() string {
	return `ip netns list-id`
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "list network namespace identifiers",
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
	_, args = options.New(args)
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}
	namebyid := make(map[int32]string)
	sock, err := nl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	sr := nl.NewSockReceiver(sock)

	varRunNetns, err := ioutil.ReadDir(rtnl.VarRunNetns)
	if err == nil {
		for _, fi := range varRunNetns {
			nsid, err := rtnl.Nsid(sr, fi.Name())
			if err == nil && nsid >= 0 {
				namebyid[nsid] = fi.Name()
			}
		}
	}
	req, err := nl.NewMessage(
		nl.Hdr{
			Type:  rtnl.RTM_GETNSID,
			Flags: nl.NLM_F_REQUEST | nl.NLM_F_DUMP,
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
		t := nl.HdrPtr(b).Type
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
			nsid = nl.Int32(val)
		}
		fmt.Print(nsid)
		if name, found := namebyid[nsid]; found {
			fmt.Print(": ", name)
		}
		fmt.Println()
	})
}
