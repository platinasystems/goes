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

var explanations = map[dnsmessage.Type]string{
	dnsmessage.TypeA:     "has address",
	dnsmessage.TypeNS:    "name server",
	dnsmessage.TypeCNAME: "is an alias for",
	dnsmessage.Type(11):  "has well known services",
	dnsmessage.TypePTR:   "domain name pointer",
	dnsmessage.Type(13):  "host information",
	dnsmessage.TypeMX:    "mail is handled by",
	dnsmessage.TypeTXT:   "descriptive text",
	dnsmessage.Type(19):  "x25 address",
	dnsmessage.Type(20):  "ISDN address",
	dnsmessage.Type(24):  "has signature",
	dnsmessage.Type(25):  "has key",
	dnsmessage.TypeAAAA:  "has IPv6 address",
	dnsmessage.Type(29):  "location",
}

func Host(ctx context.Context, args []string) error {
	var (
		name xdnsmessage.Name
		msg  dnsmessage.Message
		qhdr dnsmessage.Header
		rbuf []byte
	)
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
	} else if err = name.UnmarshalText([]byte(args[0])); err != nil {
		return xerrors.Label(err, "name")
	} else {
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
		types[0] = xdnsmessage.Type(dnsmessage.TypeALL)
	} else if flags.t == xdnsmessage.Type(dnsmessage.TypeA) {
		types = append(types, xdnsmessage.Type(dnsmessage.TypeAAAA),
			xdnsmessage.Type(dnsmessage.TypeMX))
	}
	for _, t := range types {
		qhdr.ID = xdnsmessage.NewID()
		qhdr.RecursionDesired = !flags.r
		q, err := xdnsmessage.
			NewQuestion(buf[2:], qhdr, name, t, flags.c)
		if err != nil {
			return xerrors.Mark(err)
		}
		rbuf, err = xdnsmessage.
			TimeLimitedAsk(ctx, udp, q, 30*time.Second)
		if err != nil {
			return xerrors.Mark(err)
		} else if err = msg.Unpack(rbuf); err != nil {
			return xerrors.Mark(err)
		} else if msg.ID != qhdr.ID {
			return fmt.Errorf("id %d != %d", msg.ID, qhdr.ID)
		}
		for _, r := range msg.Answers {
			fmt.Print(name, " ")
			s, ok := explanations[r.Header.Type]
			if ok {
				fmt.Print(s, " ")
			}
			fmt.Println(xdnsmessage.AnswerString(r))
		}
	}
	return nil
}
