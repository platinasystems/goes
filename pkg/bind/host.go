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
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

func Host(ctx context.Context, args []string) error {
	var name string

	xflag.TemplateUsage(`
usage: {{.Name}} [-flags] {name} [server]
Mimic BIND9's DNS lookup utility.

{{flags .}}`)

	defineHostFlags()

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

	if host_V {
		if mm := xprogram.MainModule(); mm != nil {
			fmt.Println(mm.Version)
		} else {
			fmt.Println("(unavailable)")
		}
		return nil
	}
	if host_v {
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

	types := []xdnsmessage.Type{host_t}
	if host_a || host_A {
		types[0] = xdnsmessage.TypeANY
	} else if host_t == xdnsmessage.TypeA {
		types = append(types, xdnsmessage.TypeAAAA, xdnsmessage.TypeMX)
	}
	var hf xdnsmessage.HF
	if !host_r {
		hf |= xdnsmessage.HFRecursionDesired
	}
	for _, t := range types {
		var rsp xdnsmessage.Message
		req := xdnsmessage.Message{
			HF:     hf,
			OpCode: xdnsmessage.OpCodeQuery,
			Questions: []xdnsmessage.WireQuestion{{
				Name:  xdnsmessage.MakeUniqueString(name),
				Class: host_c,
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

var (
	// Host Flags
	host_4 = false
	host_6 = false
	host_A = false
	host_C = false
	host_N = 0
	host_R = 3
	host_T = false
	host_U = true
	host_V = false
	host_W = 30 * time.Second
	host_a = false
	host_c = xdnsmessage.ClassINET
	host_i = false
	host_l = false
	host_m = false
	host_p = 53
	host_r = false
	host_s = false
	host_t = xdnsmessage.TypeA
	host_v = false
	host_w = false
)

func defineHostFlags() {
	xflag.Define(&host_4, "4", "Only use IPv4 query transport.")
	xflag.Define(&host_6, "6", "Only use IPv6 query transport.")
	xflag.Define(&host_A, "A", "Like -a but omits RRSIG, NSEC, NSEC3")
	xflag.Define(&host_C, "C",
		"Compare SOA records on authoritative servers.")
	xflag.Define(&host_N, "N",
		"Number of dots before root lookup is done.")
	xflag.Define(&host_R, "R", "UDP retries.")
	xflag.Define(&host_T, "T", "TCP mode.")
	xflag.Define(&host_U, "U", "UDP mode.")
	xflag.Define(&host_V, "V", "Print version number and exit.")
	xflag.Define(&host_W, "W", "Reply wait time.")
	xflag.Define(&host_a, "a", "Equivalent to -v -t ANY")
	xflag.Define(&host_c, "c", "Query class for non-IN data")
	xflag.Define(&host_i, "i", "FIXME?")
	xflag.Define(&host_l, "l", "Using AXFR, lists all hosts in a domain.")
	xflag.Define(&host_m, "m", "Memory debugging (trace|record|usage).")
	xflag.Define(&host_p, "p", "Server port.")
	xflag.Define(&host_r, "r", "Disable recursive processing.")
	xflag.Define(&host_s, "s", "A SERVFAIL response should stop query.")
	xflag.Define(&host_t, "t", "Query type.")
	xflag.Define(&host_v, "v", "Verbose output.")
	xflag.Define(&host_v, "d", "Debug (aka. Verbose).")
	xflag.Define(&host_w, "w", "Wait forever for a reply.")
}
