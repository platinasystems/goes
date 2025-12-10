// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package host

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/platinasystems/goes/v2/pkg/cert"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsdoh"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnspkt"
)

const HostUsage = `
usage: {{.Name}} [flags] {name} [server]
Mimic BIND9's DNS lookup utility.

{{flags .}}`

var (
	Host_4,
	Host_6,
	Host_A,
	Host_C,
	Host_S,
	Host_T,
	Host_V,
	Host_a,
	Host_i,
	Host_l,
	Host_m,
	Host_r,
	Host_s,
	Host_w bool
	Host_U = true
	Host_N int
	Host_R = 3
	Host_W = 30 * time.Second
	Host_c = xdnsmessage.ClassINET
	Host_p = 53
	Host_t = xdnsmessage.TypeA
)

var HostFlags = xflag.Labels{
	xlog.VerboseFlag,
	xmain.ConfigFlag,
	xmain.StateFlag,
	cert.VerifyFlag,
	xdnsdoh.ConfigFlag,
	{"4", "Only use IPv4 query transport.", &Host_4},
	{"6", "Only use IPv6 query transport.", &Host_6},
	{"A", "Like -a but omits RRSIG, NSEC, NSEC3.", &Host_A},
	{"C", `Compare SOA records on authoritative servers.`[1:], &Host_C},
	{"N", "Number of dots before root lookup is done.", &Host_N},
	{"R", "UDP retries.", &Host_R},
	{"S", "Skip DOH server verification. (or. -ssl-verify=false)", &Host_S},
	{"T", "TCP mode.", &Host_T},
	{"U", "UDP mode.", &Host_U},
	{"V", "Print version number and exit.", &Host_V},
	{"W", "Reply wait time.", &Host_W},
	{"a", "Equivalent to -v -t ANY", &Host_a},
	{"c", "Query class for non-IN data.", &Host_c},
	{"i", "FIXME?", &Host_i},
	{"l", "Using AXFR, lists all hosts in a domain.", &Host_l},
	{"m", "Memory debugging.", &Host_m},
	{"p", "Server port.", &Host_p},
	{"r", "Disable recursive processing.", &Host_r},
	{"s", "Stop query on SERVFAIL response.", &Host_s},
	{"t", "Query type.", &Host_t},
	{"w", "Wait forever for a reply.", &Host_w},
}

func Host(ctx context.Context, args []string) error {
	const class = xdnsmessage.ClassINET
	var name string
	var rsp xdnsmessage.Message

	xflag.TemplateUsage(HostUsage)
	err := HostFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	if Host_V {
		fmt.Println(xmain.Version())
		return nil
	}

	svr := "localhost"

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

	b := xdnsmessage.MakeBuffer()

	ask, err := xdnsdoh.Asker()
	if err != nil {
		return err
	} else if ask != nil {
		name = xdnsdoh.FQDN(name)
	} else {
		nw := "udp"
		if Host_4 {
			nw = "udp4"
		} else if Host_6 {
			nw = "udp6"
		}
		udp, err := xdns.DialContext(ctx, nw, svr)
		if err != nil {
			return err
		}
		defer udp.Close()
		ask = xdnspkt.NewTimeLimitedAsk(udp, 30*time.Second)
	}

	types := []xdnsmessage.Type{Host_t}
	if Host_a || Host_A {
		types[0] = xdnsmessage.TypeANY
	} else if Host_t == xdnsmessage.TypeA {
		types = append(types, xdnsmessage.TypeAAAA, xdnsmessage.TypeMX)
	}
	us := xdnsmessage.MakeUniqueString(name)
	for _, t := range types {
		q := xdnsmessage.NewQuery(!Host_r, us, class, t)
		b, err = q.AppendTo(b[:0])
		if err != nil {
			return err
		}
		if b, err = ask(ctx, b); err != nil {
			return err
		}
		if err = rsp.UnmarshalBinary(b); err != nil {
			return err
		}
		if rsp.ID != q.ID {
			return fmt.Errorf("id %d != %d", rsp.ID, q.ID)
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
