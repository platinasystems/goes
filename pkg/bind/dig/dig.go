// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package dig

import (
	"bufio"
	"context"
	_ "embed"
	"flag"
	"fmt"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsdoh"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnspkt"
	"github.com/platinasystems/goes/v2/pkg/xnet/xresolv"
)

const (
	DefaultAdditional = true
	DefaultAnswer     = true
	DefaultAuthority  = true
	DefaultCmd        = true
	DefaultComments   = true
	DefaultNdots      = 1
	DefaultQuestion   = true
	DefaultRecurse    = true
	DefaultStats      = true
)

//go:embed usage.txt
var usage string

var r *xresolv.Resolv

var allq = struct {
	dns xdns.Asker
	buf []byte

	ip4, ip6, m, r, u bool

	p uint16

	svr, b, f, k, y string

	o options
}{
	p: 53,
	o: options{
		additional: DefaultAdditional,
		answer:     DefaultAnswer,
		authority:  DefaultAuthority,
		cmd:        DefaultCmd,
		comments:   DefaultComments,
		ndots:      DefaultNdots,
		question:   DefaultQuestion,
		recurse:    DefaultRecurse,
		stats:      DefaultStats,
	},
}

func Dig(ctx context.Context, args []string) error {
	var err error
	xflag.TemplateUsage(usage)
	if len(args) == 0 {
		flag.CommandLine.Usage()
		return xerrors.ErrIncomplete
	}
	switch args[0] {
	case "-h", "help", "-help", "--help":
		flag.CommandLine.Usage()
		return nil
	case "-v":
		fmt.Println(xmain.Version())
		return nil
	}

	allq.buf = xdnsmessage.MakeBuffer()
	r = xresolv.New()

	cmd := strings.Join(args, " ")
	if strings.HasPrefix(args[0], "@") {
		allq.svr = strings.TrimPrefix(args[0], "@")
		if args = args[1:]; len(args) == 0 {
			return xerrors.ErrIncomplete
		}
	} else {
		allq.svr = r.Nameserver[0]
	}

	for {
		if len(args) == 0 {
			err = xerrors.ErrIncomplete
		} else if args[0] == "-f" {
			if len(args) > 1 {
				allq.f = args[1]
				args = args[2:]
			} else {
				err = xerrors.Incomplete("filename")
			}
		} else if args[0] == "-q" {
			break
		} else if args[0] == "-t" {
			break
		} else if args[0] == "-c" {
			break
		} else if args[0] == "-x" {
			break
		} else if args[0] == "-4" {
			allq.ip4 = true
			args = args[1:]
		} else if args[0] == "-6" {
			allq.ip6 = true
			args = args[1:]
		} else if args[0] == "-m" {
			allq.m = true
			args = args[1:]
		} else if args[0] == "-p" {
			if len(args) > 1 {
				_, err = fmt.Sscan(args[1], &allq.p)
				args = args[2:]
			} else {
				err = xerrors.Incomplete("port")
			}
		} else if args[0] == "-r" {
			allq.r = true
			args = args[1:]
		} else if args[0] == "-u" {
			allq.u = true
			args = args[1:]
		} else if strings.HasPrefix(args[0], "-") {
			err = xerrors.Invalid(args[0])
		} else if strings.HasPrefix(args[0], "+no") {
			err = xerrors.Invalid(args[0])
		} else if strings.HasPrefix(args[0], "+") {
			err = allq.o.set(strings.TrimPrefix(args[0], "+"))
		} else {
			break
		}
		if err != nil {
			return err
		}
	}
	if len(allq.o.https) > 0 {
		allq.dns = xdnsdoh.New(allq.o.httpsSkipVerify, allq.svr)
	} else {
		var a string
		nw := "udp"
		if allq.ip4 {
			nw = "udp4"
		} else if allq.ip6 {
			nw = "udp6"
		}
		if strings.IndexRune(allq.svr, ':') >= 0 {
			a = fmt.Sprint("[", allq.svr, "]:", allq.p)
		} else {
			a = fmt.Sprint(allq.svr, ":", allq.p)
		}
		udp, err := xdns.DialContext(ctx, nw, a)
		if err != nil {
			return err
		}
		defer udp.Close()
		allq.svr += fmt.Sprint("(", udp.RemoteAddr(), ")")
		allq.dns = xdnspkt.TimeLimitedAsker(udp, 30*time.Second)
	}
	if !allq.o.short && allq.o.cmd {
		fmt.Println("; <<>> goes-dig", xmain.Version(), "<<>>", cmd)
		fmt.Println()
		fmt.Println(";; global options:", &allq.o)
	}
	if len(allq.f) > 0 {
		err = batch(ctx)
	} else {
		for err == nil && len(args) > 0 {
			args, err = query(ctx, args)
		}
	}
	return err
}

func batch(ctx context.Context) error {
	var err error
	var sc *bufio.Scanner
	if allq.f == "-" {
		sc = bufio.NewScanner(os.Stdin)
	} else if f, err := os.Open(allq.f); err != nil {
		return err
	} else {
		defer f.Close()
		sc = bufio.NewScanner(f)
	}
	for err == nil && sc.Scan() {
		line := sc.Text()
		if len(line) == 0 ||
			strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, ";") {
			continue
		}
		args := strings.Fields(line)
		for err == nil && len(args) > 0 {
			args, err = query(ctx, args)
		}
	}
	return err
}

// allq.dns.Ask with parsed args and return remainder
func query(ctx context.Context, args []string) ([]string, error) {
	var err error
	var rsp xdnsmessage.Message
	var s string
	c := xdnsmessage.ClassINET
	t := xdnsmessage.TypeA
	o := allq.o
	for len(args) > 0 && err == nil {
		if args[0] == "--" || args[0] == "++" {
			break
		} else if args[0] == "-q" {
			if len(args) < 2 {
				err = xerrors.Incomplete("name")
			} else {
				s = o.expandedName(args[1])
				args = args[2:]
			}
		} else if args[0] == "-c" {
			if len(args) < 2 {
				err = xerrors.Incomplete("class")
			} else if c, err = xdnsmessage.
				ClassNamed(args[0]); err == nil {
				args = args[2:]
			}
		} else if args[0] == "-t" {
			if len(args) < 2 {
				err = xerrors.Incomplete("type")
			} else if t, err = xdnsmessage.
				TypeNamed(args[0]); err == nil {
				args = args[2:]
			}
		} else if args[0] == "-x" {
			var a netip.Addr
			if len(args) < 2 {
				err = xerrors.Incomplete("address")
			} else if a, err = netip.
				ParseAddr(args[1]); err == nil {
				s = xdnsmessage.Reverse(a)
				t = xdnsmessage.TypePTR
				c = xdnsmessage.ClassINET
				args = args[2:]
			}
		} else if strings.HasPrefix(args[0], "+no") {
			name := strings.TrimPrefix(args[0], "+no")
			if err = o.reset(name); err == nil {
				args = args[1:]
			}
		} else if strings.HasPrefix(args[0], "+") {
			name := strings.TrimPrefix(args[0], "+")
			if err = o.set(name); err == nil {
				args = args[1:]
			}
		} else if v, e := xdnsmessage.TypeNamed(args[0]); e == nil {
			t = v
			args = args[1:]
		} else if v, e := xdnsmessage.ClassNamed(args[0]); e == nil {
			c = v
			args = args[1:]
		} else if len(s) > 0 {
			break
		} else {
			s = o.expandedName(args[0])
			args = args[1:]
		}
	}
	if err != nil {
		return args, err
	}
	if len(s) == 0 {
		return args, xerrors.Incomplete("name")
	}
	us := xdnsmessage.MakeUniqueString(s)
	q := xdnsmessage.NewQuery(o.recurse, us, c, t)
	if allq.buf, err = q.AppendTo(allq.buf[:0]); err != nil {
		return args, err
	}
	beg := time.Now()
	if allq.buf, err = allq.dns.Ask(ctx, allq.buf); err != nil {
		return args, err
	}
	end := time.Now()
	if err = rsp.UnmarshalBinary(allq.buf); err != nil {
		return args, err
	}
	if rsp.ID != q.ID {
		return args, fmt.Errorf("id %d != %d", rsp.ID, q.ID)
	}
	if !o.short && o.comments {
		fmt.Println(";; Got answer:")
		fmt.Print(";; ->>HEADER<<- opcode: ", rsp.OpCode)
		fmt.Print(", status: ", rsp.RCode)
		fmt.Printf(", id: %d", rsp.ID)
		fmt.Println()
		fmt.Printf(";; flags: %s", rsp.HF)
		fmt.Printf("; QUERY: %d", len(rsp.Questions))
		fmt.Printf(", ANSWER: %d", len(rsp.Answers))
		fmt.Printf(", AUTHORITY: %d", len(rsp.Authorities))
		fmt.Printf(", ADDITIONAL: %d", len(rsp.Additionals))
		fmt.Println()
		fmt.Println()
	}
	if !o.short && o.additional && len(rsp.Additionals) > 0 {
		if o.comments {
			fmt.Println(";; OPT PSEUDOSECTION:")
		}
		for _, a := range rsp.Additionals {
			edns := a.Seconds()
			fmt.Print("; EDNS: version: ",
				xdnsmessage.EDNSVersion(edns))
			if xdnsmessage.HasEDNS0DNSSECOK(edns) {
				fmt.Print(" do")
			}
			mbz := xdnsmessage.EDNS0MBZ(edns)
			n := uint16(a.Class)
			if mbz != 0 {
				fmt.Printf("; MBZ: %#.4x, udp: %d\n", mbz, n)
			} else {
				fmt.Println("; udp:", n)
			}
		}
	}
	if !o.short && o.question && len(rsp.Questions) > 0 {
		if o.comments {
			fmt.Println(";; QUESTION SECTION:")
		}
		for _, q := range rsp.Questions {
			fmt.Printf(";%-31s", q.Name)
			fmt.Printf("%-8s", q.Class)
			fmt.Printf("%-s\n", q.Type)
		}
		if o.comments {
			fmt.Println()
		}
	}
	if o.answer && len(rsp.Answers) > 0 {
		if !o.short && o.comments {
			fmt.Println(";; ANSWER SECTION:")
		}
		for _, a := range rsp.Answers {
			if o.short {
				fmt.Print(a)
				continue
			}
			fmt.Printf("%-24s", a.Name)
			fmt.Printf("%-8d", a.Seconds())
			fmt.Printf("%-8s", a.Class)
			fmt.Printf("%-8s", a.Type())
			xdnsmessage.LineWrap(os.Stdout, a.String(), 24+8+8+8)
		}
		if o.comments {
			fmt.Println()
		}
	}
	if !o.short && o.authority && len(rsp.Authorities) > 0 {
		if o.comments {
			fmt.Println(";; AUTHORITY SECTION:")
		}
		for _, a := range rsp.Authorities {
			if o.short {
				fmt.Print(a)
				continue
			}
			fmt.Printf("%-24s", a.Name)
			fmt.Printf("%-8d", a.Seconds())
			fmt.Printf("%-8s", a.Class)
			fmt.Printf("%-8s", a.Type())
			xdnsmessage.LineWrap(os.Stdout, a.String(), 24+8+8+8)
		}
		if o.comments {
			fmt.Println()
		}
	}
	if !o.short && o.stats {
		ef := end.Format("Mon Jan 01 15:04:05 MST 2006")
		fmt.Print(";; Query time: ")
		dur := end.Sub(beg)
		if allq.u {
			fmt.Println(dur.Microseconds(), "µsec")
		} else {
			fmt.Println(dur.Milliseconds(), "msec")
		}
		fmt.Println(";; SERVER: ", allq.svr)
		fmt.Println(";; WHEN:", ef)
		fmt.Println(";; MSG SIZE:", len(allq.buf))
	}
	return args, nil
}
