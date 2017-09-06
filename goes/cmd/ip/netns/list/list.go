// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package list

import (
	"fmt"
	"io/ioutil"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "list"
	Apropos = "list network namespaces"
	Usage   = `ip netns [ list ]`
	Man     = `
SEE ALSO
	ip man netns || ip netns -man
	man ip || ip -man`
)

var man = lang.Alt{
	lang.EnUS: Man,
}

func New(s string) Command { return Command(s) }

type Command string

func (Command) Aka() string { return "list" }

func (c Command) Apropos() lang.Alt {
	apropos := Apropos
	if c == "list" {
		apropos += " (default)"
	}
	return lang.Alt{
		lang.EnUS: apropos,
	}
}

func (Command) Man() lang.Alt    { return man }
func (c Command) String() string { return string(c) }
func (Command) Usage() string    { return Usage }

func (Command) Main(args ...string) error {
	_, args = options.New(args)
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}
	varRunNetns, err := ioutil.ReadDir(rtnl.VarRunNetns)
	if err != nil {
		return err
	}
	sock, err := rtnl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	sr := rtnl.NewSockReceiver(sock)

	for _, fi := range varRunNetns {
		fmt.Print(fi.Name())
		nsid, err := sr.Nsid(fi.Name())
		if err == nil && nsid >= 0 {
			fmt.Print(": ", nsid)
		}
		fmt.Println()
	}
	return nil
}
