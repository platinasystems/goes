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

var host_4, host_6, host_A, host_C, host_S, host_T, host_V, host_a, host_i,
	host_l, host_m, host_r, host_s, host_w bool
var host_U = true
var host_N int
var host_R = 3
var host_W = 30 * time.Second
var host_c = xdnsmessage.ClassINET
var host_p = 53
var host_t = xdnsmessage.TypeA

var hostFlags = xflag.Labels{
	xlog.VerboseFlag,
	xmain.ConfigFlag,
	xmain.StateFlag,
	cert.VerifyFlag,
	xdnsdoh.ConfigFlag,
	{"4", "Only use IPv4 query transport.", &host_4},
	{"6", "Only use IPv6 query transport.", &host_6},
	{"A", "Like -a but omits RRSIG, NSEC, NSEC3.", &host_A},
	{"C", `Compare SOA records on authoritative servers.`[1:], &host_C},
	{"N", "Number of dots before root lookup is done.", &host_N},
	{"R", "UDP retries.", &host_R},
	{"S", "Skip DOH server verification. (or. -ssl-verify=false)", &host_S},
	{"T", "TCP mode.", &host_T},
	{"U", "UDP mode.", &host_U},
	{"V", "Print version number and exit.", &host_V},
	{"W", "Reply wait time.", &host_W},
	{"a", "Equivalent to -v -t ANY", &host_a},
	{"c", "Query class for non-IN data.", &host_c},
	{"i", "FIXME?", &host_i},
	{"l", "Using AXFR, lists all hosts in a domain.", &host_l},
	{"m", "Memory debugging.", &host_m},
	{"p", "Server port.", &host_p},
	{"r", "Disable recursive processing.", &host_r},
	{"s", "Stop query on SERVFAIL response.", &host_s},
	{"t", "Query type.", &host_t},
	{"w", "Wait forever for a reply.", &host_w},
}

func Host(ctx context.Context, args []string) error {
	const class = xdnsmessage.ClassINET

	xflag.TemplateUsage(`
usage: {{.Name}} [-flags] {name} [server]
Mimic BIND9's DNS lookup utility.

{{flags .}}`)

	var name string
	var rsp xdnsmessage.Message

	err := hostFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	if host_V {
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

	ask, err := xdnsdoh.Asker(false)
	if err != nil {
		return err
	} else if ask != nil {
		name = xdnsdoh.FQDN(name)
	} else {
		nw := "udp"
		if host_4 {
			nw = "udp4"
		} else if host_6 {
			nw = "udp6"
		}
		udp, err := xdns.DialContext(ctx, nw, svr)
		if err != nil {
			return err
		}
		defer udp.Close()
		ask = xdnspkt.NewTimeLimitedAsk(udp, 30*time.Second)
	}

	types := []xdnsmessage.Type{host_t}
	if host_a || host_A {
		types[0] = xdnsmessage.TypeANY
	} else if host_t == xdnsmessage.TypeA {
		types = append(types, xdnsmessage.TypeAAAA, xdnsmessage.TypeMX)
	}
	us := xdnsmessage.MakeUniqueString(name)
	for _, t := range types {
		q := xdnsmessage.NewQuery(!host_r, us, class, t)
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
