// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package host

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
	"golang.org/x/net/dns/dnsmessage"
)

var (
	mutable = log.New(os.Stdout, "", log.Lshortfile)
	errata  = xlog.Unmute(mutable)
	verbose = xlog.Mute(mutable)
)

func Host(ctx context.Context, args []string) error {
	var name xdns.Name
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
	udp, err := xdns.NewUDP(ctx, svr, flags.p)
	if err != nil {
		return err
	}
	defer udp.Close()
	qbuf := make([]byte, 2, 2+512)
	qhdr := dnsmessage.Header{
		ID: xdns.NewID(),
	}
	if !flags.r {
		qhdr.RecursionDesired = true
	}
	q, err := xdns.NewQuestion(qbuf[2:], qhdr, name.Name, flags.t, flags.c)
	if err != nil {
		return err
	}
	msg, _, err := xdns.AskUDP(ctx, udp, q)
	if err != nil {
		return err
	}
	if msg.ID != qhdr.ID {
		return fmt.Errorf("id %d != %d", msg.ID, qhdr.ID)
	}
	for _, r := range msg.Answers {
		switch r.Header.Type {
		case dnsmessage.TypeA:
			ŕ := r.Body.(*dnsmessage.AResource)
			fmt.Print(net.IP(ŕ.A[:]))
		case dnsmessage.TypeNS:
			ŕ := r.Body.(*dnsmessage.NSResource)
			fmt.Print(ŕ.NS)
		case dnsmessage.TypeCNAME:
			ŕ := r.Body.(*dnsmessage.CNAMEResource)
			fmt.Print(ŕ.CNAME)
		case dnsmessage.TypeSOA:
			ŕ := r.Body.(*dnsmessage.SOAResource)
			fmt.Printf("ns %v, mbox %v, s/n %d",
				ŕ.NS, ŕ.MBox, ŕ.Serial)
		case dnsmessage.TypePTR:
			ŕ := r.Body.(*dnsmessage.PTRResource)
			fmt.Print(ŕ.PTR)
		case dnsmessage.TypeMX:
			ŕ := r.Body.(*dnsmessage.MXResource)
			fmt.Printf("%v, pref %d", ŕ.MX, ŕ.Pref)
		case dnsmessage.TypeTXT:
			ŕ := r.Body.(*dnsmessage.TXTResource)
			fmt.Print(strings.Join(ŕ.TXT, " "))
		case dnsmessage.TypeAAAA:
			ŕ := r.Body.(*dnsmessage.AAAAResource)
			fmt.Print(net.IP(ŕ.AAAA[:]))
		case dnsmessage.TypeSRV:
			ŕ := r.Body.(*dnsmessage.SRVResource)
			fmt.Printf("%v, port %d, pri %d, weight %d",
				ŕ.Target, ŕ.Port, ŕ.Priority, ŕ.Weight)
		}
		fmt.Println()
	}
	return nil
}
