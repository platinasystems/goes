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
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnspkt"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
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
	svr := "localhost:domain"

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

	pkt := xdnspkt.Pool.Alloc(0)
	defer xdnspkt.Pool.Free(pkt)

	rsvp := func(data []byte) ([]byte, error) {
		return data, xerrors.Incomplete("requester")
	}
	if strings.HasPrefix(svr, "https:") {
		return xerrors.FIXME("DOH")
	} else {
		udp, err := xdns.DialContext(ctx, "udp", svr)
		if err != nil {
			return err
		}
		defer udp.Close()
		rsvp = func(b []byte) ([]byte, error) {
			const tl = 30 * time.Second
			return xdnspkt.TimeLimitedAsk(ctx, udp, b, tl)
		}
	}

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
		var rsp xdnsmessage.Message
		req := xdnsmessage.Message{
			HF:     hf,
			OpCode: xdnsmessage.OpCodeQuery,
			Questions: []xdnsmessage.WireQuestion{{
				Name:  xdnsmessage.MakeUniqueString(name),
				Class: flags.c,
				Type:  t,
			}},
		}
		if pkt, err = req.AppendTo(pkt[:0]); err != nil {
			return err
		}
		if pkt, err = rsvp(pkt); err != nil {
			return err
		}
		if err = rsp.UnmarshalBinary(pkt); err != nil {
			return err
		}
		if rsp.ID != req.ID {
			return fmt.Errorf("id %d != %d", rsp.ID, req.ID)
		}
		for _, a := range rsp.Answers {
			fmt.Print(name, " ")
			s, ok := explanations[a.Type()]
			if ok {
				fmt.Print(s, " ")
			}
			fmt.Println(a)
		}
	}
	return nil
}
