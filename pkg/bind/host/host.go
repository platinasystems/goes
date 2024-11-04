// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package host

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
	"golang.org/x/net/dns/dnsmessage"
)

var (
	mutable = log.New(os.Stdout, "", log.Lshortfile)
	errata  = xlog.Unmute(mutable)
	verbose = xlog.Mute(mutable)
)

var explanations = map[xdnsmessage.Type]string{
	xdnsmessage.TypeA:     "has address",
	xdnsmessage.TypeNS:    "name server",
	xdnsmessage.TypeCNAME: "is an alias for",
	xdnsmessage.TypeWKS:   "has well known services",
	xdnsmessage.TypePTR:   "domain name pointer",
	xdnsmessage.TypeHINFO: "host information",
	xdnsmessage.TypeMX:    "mail is handled by",
	xdnsmessage.TypeTXT:   "descriptive text",
	xdnsmessage.TypeX25:   "x25 address",
	xdnsmessage.TypeISDN:  "ISDN address",
	xdnsmessage.TypeSIG:   "has signature",
	xdnsmessage.TypeKEY:   "has key",
	xdnsmessage.TypeAAAA:  "has IPv6 address",
	xdnsmessage.TypeLOC:   "location",
}

func Host(ctx context.Context, args []string) error {
	var name string

	buf := make([]byte, 2, 2+xdnsmessage.MaxPacketSize)
	svr := "127.0.0.1"

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [-flags] {name} [server]
Mimic BIND9's DNS lookup utility.

{{flags .}}`)

	addCommandLineFlags()
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	if flags.V {
		if mm := xprogram.MainModule(); mm != nil {
			fmt.Println(mm.Version)
		} else {
			fmt.Println("(unavailable)")
		}
		return nil
	}
	if flags.v {
		verbose = xlog.Unmute(verbose)
	}
	args = flag.CommandLine.Args()
	if len(args) == 0 {
		return xerrors.Incomplete("name")
	} else {
		name = args[0]
		args = args[1:]
	}
	if len(args) > 0 {
		svr = args[0]
	}
	// FIXME alt DOH
	udp, err := xdnsmessage.NewUDP(ctx, svr, flags.p)
	if err != nil {
		return err
	}
	defer udp.Close()
	types := []xdnsmessage.Type{flags.t}
	if flags.a || flags.A {
		types[0] = xdnsmessage.TypeANY
	} else if flags.t == xdnsmessage.TypeA {
		types = append(types, xdnsmessage.TypeAAAA, xdnsmessage.TypeMX)
	}
	var hf xdnsmessage.HF
	if !flags.r {
		hf |= xdnsmessage.HFRecursionDesired
	}
	for _, t := range types {
		var msg dnsmessage.Message

		id, q, err := xdnsmessage.
			NewQuestion(buf[2:], name, t, flags.c, hf)
		if err != nil {
			return xerrors.Mark(err)
		}
		data, err := xdnsmessage.
			TimeLimitedAsk(ctx, udp, q, 30*time.Second)
		if err != nil {
			return xerrors.Mark(err)
		}
		if err = msg.Unpack(data); err != nil {
			return xerrors.Mark(err)
		}
		if msg.ID != id {
			return fmt.Errorf("id %d != %d", msg.ID, id)
		}
		for _, r := range msg.Answers {
			fmt.Print(name, " ")
			s, ok := explanations[xdnsmessage.Type(r.Header.Type)]
			if ok {
				fmt.Print(s, " ")
			}
			fmt.Println(xdnsmessage.AnswerString(r))
		}
	}
	return nil
}
