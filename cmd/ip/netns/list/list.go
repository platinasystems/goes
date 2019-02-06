// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package list

import (
	"fmt"
	"io/ioutil"

	"github.com/platinasystems/goes/cmd/ip/internal/options"
	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/goes/internal/nl"
	"github.com/platinasystems/goes/internal/nl/rtnl"
)

type Command string

func (Command) Aka() string { return "list" }

func (c Command) String() string { return string(c) }

func (Command) Usage() string {
	return `ip netns [ list ]`
}

func (c Command) Apropos() lang.Alt {
	apropos := "list network namespaces"
	if c == "list" {
		apropos += " (default)"
	}
	return lang.Alt{
		lang.EnUS: apropos,
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
	varRunNetns, err := ioutil.ReadDir(rtnl.VarRunNetns)
	if err != nil {
		return err
	}
	sock, err := nl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	sr := nl.NewSockReceiver(sock)

	for _, fi := range varRunNetns {
		fmt.Print(fi.Name())
		nsid, err := rtnl.Nsid(sr, fi.Name())
		if err == nil && nsid >= 0 {
			fmt.Print(": ", nsid)
		}
		fmt.Println()
	}
	return nil
}
