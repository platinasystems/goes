// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package add

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "add"
	Apropos = "virtual link"
	Usage   = `
	ip link add [ link DEVICE ] [ name ] NAME
		[ txqueuelen PACKETS ] [ address LLADDR ] [ broadcast LLADDR ]
		[ mtu MTU ] [ index IDX ] [ numtxqueues QUEUE_COUNT ]
		[ numrxqueues QUEUE_COUNT ] type TYPE [ ARGS ]

	TYPE := { bridge | bond | can | dummy | hsr | ifb | ipoib | macvlan |
		macvtap | vcan | veth | vlan | vxlan | ip6tnl | ipip | sit |
		gre | gretap | ip6gre | ip6gretap | vti | nlmon | ipvlan |
		lowpan | geneve | vrf | macsec }
	`
	Man = `
SEE ALSO
	ip man link || ip link -man
`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
	Parms = []interface{}{
		"link", "name", "txqueuelen", "address", "broadcast", "mtu",
		"index", "numtxqueues", "numrxqueues", "type",
	}
)

func New() Command { return Command{} }

type Command struct{}

type add options.Options

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	var err error

	if args, err = options.Netns(args); err != nil {
		return err
	}

	o, args := options.New(args)
	add := (*add)(o)
	args = add.Parms.More(args, Parms)

	fmt.Println("FIXME", Name, args)

	return nil
}
