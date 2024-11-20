// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package dig

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnspkt"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

func DiG(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [@server] [+global] [[-flags] [name [TYPE] [CLASS] [+options]]
Mimic BIND9's DNS lookup utility.

{{flags .}}`)
	addCommandLineFlags()
	return lookup(ctx, flag.CommandLine, nil, args)
}

func lookup(
	ctx context.Context,
	fs *flag.FlagSet,
	rsvp func([]byte) ([]byte, error),
	args []string,
) error {
	const hf = xdnsmessage.HFRecursionDesired
	var (
		cmd,
		name,
		ra,
		svr string
		c   xdnsmessage.Class
		t   xdnsmessage.Type
		rsp xdnsmessage.Message
		err error
	)

	pkt := xdnspkt.Pool.Alloc(0)
	defer xdnspkt.Pool.Free(pkt)

	if fs == flag.CommandLine {
		cmd = strings.Join(args, " ")
		svr = "localhost:domain"
		if len(args) > 0 && strings.HasPrefix(args[0], "@") {
			svr = strings.TrimPrefix(args[0], "@")
			args = args[1:]
		}
	}
	args, err = gopts.parse(args)
	if err != nil {
		return err
	}
	addQueryFlags(fs)
	if err = fs.Parse(args); err != nil {
		return err
	}
	args = fs.Args()
	if fs == flag.CommandLine {
		if flags.O {
			fmt.Print(optionsTxt)
			return nil
		}
		if flags.T {
			fmt.Print(xdnsmessage.TypeHelpTxt)
			return nil
		}
		ver := "(unavailable)"
		if mm := xprogram.MainModule(); mm != nil {
			ver = mm.Version
		}
		if flags.v {
			fmt.Println(ver)
			return nil
		}
		if strings.HasPrefix(svr, "https:") {
			rsvp = func(b []byte) ([]byte, error) {
				return b[:0], xerrors.FIXME("DOH")
			}
			// FIXME defer cl.CloseIdleConnections()
			return xerrors.FIXME("DOH")
		} else {
			nw := "udp"
			if flags.ip4only {
				nw = "udp4"
			} else if flags.ip6only {
				nw = "udp6"
			}
			conn, err := xdns.DialContext(ctx, nw, svr)
			if err != nil {
				return err
			}
			defer func() {
				conn.Close()
			}()
			ra = conn.RemoteAddr().String()
			rsvp = func(b []byte) ([]byte, error) {
				const tl = 30 * time.Second
				return xdnspkt.TimeLimitedAsk(ctx, conn, b, tl)
			}
		}
		if !gopts.has(boolOptShort) && gopts.has(boolOptCmd) {
			fmt.Printf("; <<>> goes/pkg/bind/dig %s <<>> %s\n",
				ver, cmd)
			fmt.Println()
		}
		if len(flags.f) > 0 {
			return batch(ctx, rsvp, flags.f)
		}
	}
	if s := flags.x; len(s) > 0 {
		addr, err := netip.ParseAddr(s)
		if err != nil {
			return xerrors.Label(err, "x")
		}
		name = xdnsmessage.Reverse(addr)
		t = xdnsmessage.TypePTR
		c = xdnsmessage.ClassINET
	} else {
		if len(flags.q) > 0 {
			name = flags.q
		} else if len(args) == 0 {
			return xerrors.Incomplete("name")
		} else {
			name = args[0]
			args = args[1:]
		}
		if flags.t != xdnsmessage.Type0 {
			t = flags.t
		} else if len(args) == 0 {
			t = xdnsmessage.TypeA
		} else if t, err = xdnsmessage.TypeNamed(args[0]); err != nil {
			t = xdnsmessage.TypeA
		} else {
			args = args[1:]
		}
		if flags.c != xdnsmessage.Class0 {
			c = flags.c
		} else if len(args) == 0 {
			c = xdnsmessage.ClassINET
		} else if c, err = xdnsmessage.ClassNamed(args[0]); err != nil {
			c = xdnsmessage.ClassINET
		} else {
			args = args[1:]
		}
	}

	qopts := gopts.clone()
	if args, err = qopts.parse(args); err != nil {
		return err
	}

	req := xdnsmessage.Message{
		HF:     hf,
		OpCode: xdnsmessage.OpCodeQuery,
		Questions: []xdnsmessage.WireQuestion{{
			Name:  xdnsmessage.MakeUniqueString(name),
			Class: c,
			Type:  t,
		}},
	}
	if pkt, err = req.AppendTo(pkt[:0]); err != nil {
		return err
	}
	beg := time.Now()
	if pkt, err = rsvp(pkt); err != nil {
		return err
	}
	end := time.Now()
	if err = rsp.UnmarshalBinary(pkt); err != nil {
		return err
	}
	if rsp.ID != req.ID {
		return fmt.Errorf("id %d != %d", rsp.ID, req.ID)
	}
	if !qopts.has(boolOptShort) && qopts.has(boolOptComments) {
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
	if !qopts.has(boolOptShort) && qopts.has(boolOptAdditional) &&
		len(rsp.Additionals) > 0 {
		if qopts.has(boolOptComments) {
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
	if !qopts.has(boolOptShort) && qopts.has(boolOptQuestion) &&
		len(rsp.Questions) > 0 {
		if qopts.has(boolOptComments) {
			fmt.Println(";; QUESTION SECTION:")
		}
		for _, q := range rsp.Questions {
			fmt.Printf(";%-31s", q.Name)
			fmt.Printf("%-8s", q.Class)
			fmt.Printf("%-s\n", q.Type)
		}
		if qopts.has(boolOptComments) {
			fmt.Println()
		}
	}
	if qopts.has(boolOptAnswer) && len(rsp.Answers) > 0 {
		if qopts.has(boolOptComments) && !qopts.has(boolOptShort) {
			fmt.Println(";; ANSWER SECTION:")
		}
		for _, a := range rsp.Answers {
			if qopts.has(boolOptShort) {
				fmt.Print(a)
				continue
			}
			fmt.Printf("%-24s", a.Name)
			fmt.Printf("%-8d", a.Seconds())
			fmt.Printf("%-8s", a.Class)
			fmt.Printf("%-8s", a.Type())
			xdnsmessage.LineWrap(os.Stdout, a.String(), 24+8+8+8)
		}
		if qopts.has(boolOptComments) {
			fmt.Println()
		}
	}
	if !qopts.has(boolOptShort) && qopts.has(boolOptAuthority) &&
		len(rsp.Authorities) > 0 {
		if qopts.has(boolOptComments) {
			fmt.Println(";; AUTHORITY SECTION:")
		}
		for _, a := range rsp.Authorities {
			if qopts.has(boolOptShort) {
				fmt.Print(a)
				continue
			}
			fmt.Printf("%-24s", a.Name)
			fmt.Printf("%-8d", a.Seconds())
			fmt.Printf("%-8s", a.Class)
			fmt.Printf("%-8s", a.Type())
			xdnsmessage.LineWrap(os.Stdout, a.String(), 24+8+8+8)
		}
		if qopts.has(boolOptComments) {
			fmt.Println()
		}
	}
	if !qopts.has(boolOptShort) && qopts.has(boolOptStats) {
		ef := end.Format("Mon Jan 01 15:04:05 MST 2006")
		fmt.Print(";; Query time: ")
		dur := end.Sub(beg)
		if flags.u {
			fmt.Println(dur.Microseconds(), "µsec")
		} else {
			fmt.Println(dur.Milliseconds(), "msec")
		}
		if fs == flag.CommandLine {
			fmt.Printf(";; SERVER: %s(%s)\n", svr, ra)
		}
		fmt.Println(";; WHEN:", ef)
		fmt.Println(";; MSG SIZE:", len(pkt))
	}
	if len(args) > 0 {
		return lookup(ctx, newFlagSet(), rsvp, args)
	}
	return nil
}

func batch(
	ctx context.Context,
	rsvp func([]byte) ([]byte, error),
	fn string,
) error {
	var err error
	var sc *bufio.Scanner
	if fn == "-" {
		sc = bufio.NewScanner(os.Stdin)
	} else if f, err := os.Open(fn); err != nil {
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
		err = lookup(ctx, newFlagSet(), rsvp, strings.Fields(line))
	}
	return err
}
