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
	"golang.org/x/net/dns/dnsmessage"
)

var (
	udp *net.UDPConn
	buf,
	rbuf []byte
	msg  dnsmessage.Message
	qhdr dnsmessage.Header
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
			fmt.Print(xdnsmessage.TypesTxt)
			return args, nil
		}
		if flags.v {
			fmt.Println(ver)
			return args, nil
		}
		showCmd()
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

	if s := flags.x; len(s) > 0 {
		addr, err := netip.ParseAddr(s)
		if err != nil {
			return args, xerrors.Label(err, "x")
		}
		flags.q.Reverse(addr)
		flags.t = xdnsmessage.TypePTR
		flags.c = xdnsmessage.DefaultClass
	} else if args, err = flags.q.Pull(args); err != nil {
		return args, xerrors.Label(err, "name")
	} else if args, err = flags.t.Pull(args); err != nil {
		return args, xerrors.Label(err, "type")
	} else if args, err = flags.c.Pull(args); err != nil {
		return args, xerrors.Label(err, "class")
	}

	qopts.clone(&gopts)
	args, err = qopts.parse(args)
	if err != nil {
		return args, err
	}

	qhdr.ID = xdnsmessage.NewID()
	qhdr.RecursionDesired = true
	q, err := xdnsmessage.
		NewQuestion(buf[2:], qhdr, flags.q, flags.t, flags.c)
	if err != nil {
		return args, err
	}

	beg := time.Now()
	rbuf, err = ask(ctx, q)
	end := time.Now()
	if err != nil {
		return args, err
	} else if err = msg.Unpack(rbuf); err != nil {
		return args, err
	} else if msg.ID != qhdr.ID {
		return args, fmt.Errorf("id %d != %d", msg.ID, qhdr.ID)
	}

	showHeader(&msg)
	showOptPseudoSection(msg.Additionals)
	showQuestionSection(msg.Questions)
	showAnswerSection(msg.Answers)
	showAuthoritySection(msg.Authorities)
	showStats(beg, end, len(rbuf), ra)
	return args, nil
}

func ask(ctx context.Context, b []byte) ([]byte, error) {
	if udp != nil {
		return xdnsmessage.TimeLimitedAsk(ctx, udp, b, 30*time.Second)
	}
	return b[:0], xerrors.FIXME("DOH")
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
