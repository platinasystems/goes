// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bind

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/chunk"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnspkt"
)

var Host_4 = xflag.New[bool]("4", "Only use IPv4 query transport.", nil)
var Host_6 = xflag.New[bool]("6", "Only use IPv6 query transport.", nil)
var Host_A = xflag.New[bool]("A", "Like -a but omits RRSIG, NSEC, NSEC3.", nil)
var Host_C = xflag.New[bool]("C", `
Compare SOA records on authoritative servers.`[1:],
	nil)
var Host_N = xflag.New[int]("N", "Number of dots before root lookup is done.",
	nil)
var Host_R = xflag.New[int]("R", "UDP retries.",
	func() int { return 3 })
var Host_T = xflag.New[bool]("T", "TCP mode.", nil)
var Host_U = xflag.New[bool]("U", "UDP mode.",
	func() bool { return true })
var Host_V = xflag.New[bool]("V", "Print version number and exit.", nil)
var Host_W = xflag.New[time.Duration]("W", "Reply wait time.",
	func() time.Duration { return 30 * time.Second })
var Host_a = xflag.New[bool]("a", "Equivalent to -v -t ANY", nil)
var Host_c = xflag.New[xdnsmessage.Class]("c", "Query class for non-IN data.",
	func() xdnsmessage.Class { return xdnsmessage.ClassINET })
var Host_i = xflag.New[bool]("i", "FIXME?", nil)
var Host_l = xflag.New[bool]("l", `
Using AXFR, lists all hosts in a domain.`[1:], nil)
var Host_m = xflag.New[bool]("m", "Memory debugging.", nil)
var Host_p = xflag.New[int]("p", "Server port.",
	func() int { return 53 })
var Host_r = xflag.New[bool]("r", "Disable recursive processing.", nil)
var Host_s = xflag.New[bool]("s", "Stop query on SERVFAIL response.", nil)
var Host_t = xflag.New[xdnsmessage.Type]("t", "Query type.",
	func() xdnsmessage.Type { return xdnsmessage.TypeA })
var Host_v = xflag.New[bool]("v", "Verbose output.", nil)
var Host_w = xflag.New[bool]("w", "Wait forever for a reply.", nil)

var HostFlags = []xflag.Definer{
	Host_4,
	Host_6,
	Host_A,
	Host_C,
	Host_N,
	Host_R,
	Host_T,
	Host_U,
	Host_V,
	Host_W,
	Host_a,
	Host_c,
	Host_i,
	Host_l,
	Host_m,
	Host_p,
	Host_r,
	Host_s,
	Host_t,
	Host_v,
	Host_w,
}

func Host(ctx context.Context, args []string) error {
	var name string

	xflag.TemplateUsage(`
usage: {{.Name}} [-flags] {name} [server]
Mimic BIND9's DNS lookup utility.

{{flags .}}`)

	for _, f := range HostFlags {
		f.Define()
	}

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	svr := "localhost:domain"

	explanations := map[xdnsmessage.Type]string{
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

	if Host_V.Value() {
		fmt.Println(Version())
		return nil
	}
	if Host_v.Value() {
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

	pkt := chunk.New(xdnspkt.Cap)
	*pkt = (*pkt)[:0]
	defer chunk.Discard(pkt)

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

	types := []xdnsmessage.Type{Host_t.Value()}
	if Host_a.Value() || Host_A.Value() {
		types[0] = xdnsmessage.TypeANY
	} else if Host_t.Value() == xdnsmessage.TypeA {
		types = append(types, xdnsmessage.TypeAAAA, xdnsmessage.TypeMX)
	}
	var hf xdnsmessage.HF
	if !Host_r.Value() {
		hf |= xdnsmessage.HFRecursionDesired
	}
	for _, t := range types {
		var rsp xdnsmessage.Message
		req := xdnsmessage.Message{
			HF:     hf,
			OpCode: xdnsmessage.OpCodeQuery,
			Questions: []xdnsmessage.WireQuestion{{
				Name:  xdnsmessage.MakeUniqueString(name),
				Class: Host_c.Value(),
				Type:  t,
			}},
		}
		if *pkt, err = req.AppendTo((*pkt)[:0]); err != nil {
			return err
		}
		if *pkt, err = rsvp(*pkt); err != nil {
			return err
		}
		if err = rsp.UnmarshalBinary(*pkt); err != nil {
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
