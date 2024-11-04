// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package dig

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

var (
	udp *net.UDPConn
	buf []byte
	cmd,
	svr,
	ver string
)

func DiG(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [@server] [+global] [[-flags] [name [TYPE] [CLASS] [+options]]
Mimic BIND9's DNS lookup utility.

{{flags .}}`)
	addCommandLineFlags()

	if mm := xprogram.MainModule(); mm != nil {
		ver = mm.Version
	} else {
		ver = "(unavailable)"
	}

	buf = make([]byte, 2, 2+xdnsmessage.MaxPacketSize)
	cmd = strings.Join(args, " ")
	svr = "127.0.0.1"
	if len(args) > 0 && strings.HasPrefix(args[0], "@") {
		svr = strings.TrimPrefix(args[0], "@")
		args = args[1:]
	}

	var err error
	args, err = gopts.parse(args)
	if err != nil {
		return err
	}

	// FIXME alt DOH
	defer func() {
		if udp != nil {
			udp.Close()
			udp = nil
		}
	}()

	for fs := flag.CommandLine; len(args) > 0; fs = newFlagSet() {
		if args, err = lookup(ctx, fs, args); err != nil {
			return err
		}
	}
	return nil
}

func lookup(ctx context.Context, fs *flag.FlagSet, args []string) (
	[]string, error,
) {
	var ra net.Addr
	var msg xdnsmessage.Message

	addQueryFlags(fs)
	err := fs.Parse(args)
	if err != nil {
		return args, err
	}
	args = fs.Args()
	if fs == flag.CommandLine {
		if flags.O {
			fmt.Print(optionsTxt)
			return args, nil
		}
		if flags.T {
			fmt.Print(xdnsmessage.TypeHelpTxt)
			return args, nil
		}
		if flags.v {
			fmt.Println(ver)
			return args, nil
		}
		if !gopts.has(boolOptShort) && gopts.has(boolOptCmd) {
			fmt.Printf("; <<>> goes/pkg/bind/dig %s <<>> %s\n",
				ver, cmd)
			fmt.Println()
		}
		if len(flags.f) > 0 {
			return []string{}, batch(ctx, flags.f)
		}
	}

	if udp == nil {
		// FIXME alt DOH
		udp, err = xdnsmessage.NewUDP(ctx, svr, flags.p)
		if err != nil {
			return args, err
		}
		ra = udp.RemoteAddr()
	}

	var name string
	var c xdnsmessage.Class
	var t xdnsmessage.Type

	if s := flags.x; len(s) > 0 {
		addr, err := netip.ParseAddr(s)
		if err != nil {
			return args, xerrors.Label(err, "x")
		}
		name = xdnsmessage.Reverse(addr)
		t = xdnsmessage.TypePTR
		c = xdnsmessage.ClassINET
	} else {
		if len(flags.q) > 0 {
			name = flags.q
		} else if len(args) == 0 {
			return args, xerrors.Incomplete("name")
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
	args, err = qopts.parse(args)
	if err != nil {
		return args, err
	}

	const hf = xdnsmessage.HFRecursionDesired
	id, q, err := xdnsmessage.NewQuestion(buf[2:], name, t, c, hf)
	if err != nil {
		return args, err
	}

	beg := time.Now()
	data, err := xdnsmessage.TimeLimitedAsk(ctx, udp, q, 30*time.Second)
	end := time.Now()
	if err != nil {
		return args, err
	}
	if err = msg.Unpack(data); err != nil {
		return args, err
	}
	if msg.ID != id {
		return args, fmt.Errorf("id %d != %d", msg.ID, id)
	}

	if !qopts.has(boolOptShort) && qopts.has(boolOptComments) {
		opcode := xdnsmessage.OpCode(msg.OpCode)
		rcode := xdnsmessage.RCode(msg.RCode)
		hf := xdnsmessage.NewHeaderFlags(msg.Header)
		fmt.Println(";; Got answer:")
		fmt.Print(";; ->>HEADER<<- opcode: ", opcode)
		fmt.Print(", status: ", rcode)
		fmt.Printf(", id: %d", msg.ID)
		fmt.Println()
		fmt.Printf(";; flags: %s", hf)
		fmt.Printf("; QUERY: %d", len(msg.Questions))
		fmt.Printf(", ANSWER: %d", len(msg.Answers))
		fmt.Printf(", AUTHORITY: %d", len(msg.Authorities))
		fmt.Printf(", ADDITIONAL: %d", len(msg.Additionals))
		fmt.Println()
		fmt.Println()
	}
	if !qopts.has(boolOptShort) && qopts.has(boolOptAdditional) &&
		len(msg.Additionals) > 0 {
		if qopts.has(boolOptComments) {
			fmt.Println(";; OPT PSEUDOSECTION:")
		}
		for _, r := range msg.Additionals {
			fmt.Print("; EDNS: version: ",
				xdnsmessage.EDNSVersion(r.Header.TTL))
			if xdnsmessage.HasEDNS0DNSSECOK(r.Header.TTL) {
				fmt.Print(" do")
			}
			mbz := xdnsmessage.EDNS0MBZ(r.Header.TTL)
			if mbz != 0 {
				fmt.Printf("; MBZ: %#.4x, udp: ", mbz)
			} else {
				fmt.Print("; udp: ")
			}
			fmt.Println(r.Header.Class)
		}
	}
	if !qopts.has(boolOptShort) && qopts.has(boolOptQuestion) &&
		len(msg.Questions) > 0 {
		if qopts.has(boolOptComments) {
			fmt.Println(";; QUESTION SECTION:")
		}
		for _, q := range msg.Questions {
			fmt.Printf(";%-31s", q.Name)
			fmt.Printf("%-8s", xdnsmessage.Class(q.Class))
			fmt.Printf("%-s\n", xdnsmessage.Type(q.Type))
		}
		if qopts.has(boolOptComments) {
			fmt.Println()
		}
	}
	if qopts.has(boolOptAnswer) && len(msg.Answers) > 0 {
		if qopts.has(boolOptComments) && !qopts.has(boolOptShort) {
			fmt.Println(";; ANSWER SECTION:")
		}
		for _, r := range msg.Answers {
			if !qopts.has(boolOptShort) {
				c := xdnsmessage.Class(r.Header.Class)
				t := xdnsmessage.Type(r.Header.Type)
				fmt.Printf("%-24s", r.Header.Name)
				fmt.Printf("%-8d", r.Header.TTL)
				fmt.Printf("%-8s", c)
				fmt.Printf("%-8s", t)
			}
			fmt.Println(xdnsmessage.AnswerString(r))
		}
		if qopts.has(boolOptComments) {
			fmt.Println()
		}
	}
	if !qopts.has(boolOptShort) && qopts.has(boolOptAuthority) &&
		len(msg.Authorities) > 0 {
		if qopts.has(boolOptComments) {
			fmt.Println(";; AUTHORITY SECTION:")
		}
		for _, r := range msg.Authorities {
			fmt.Println(r.Header)
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
		fmt.Printf(";; SERVER: %s#%d(%v\n", svr, flags.p, ra)
		fmt.Println(";; WHEN:", ef)
		fmt.Println(";; MSG SIZE:", len(data))
	}
	return args, nil
}

func batch(ctx context.Context, fn string) error {
	var sc *bufio.Scanner
	if fn == "-" {
		sc = bufio.NewScanner(os.Stdin)
	} else if f, err := os.Open(fn); err != nil {
		return err
	} else {
		defer f.Close()
		sc = bufio.NewScanner(f)
	}
	var err error
	for sc.Scan() {
		line := sc.Text()
		if len(line) == 0 ||
			strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, ";") {
			continue
		}
		args := strings.Fields(line)
		for err == nil && len(args) > 0 {
			args, err = lookup(ctx, newFlagSet(), args)
		}
	}
	return nil
}
