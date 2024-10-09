// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package dig

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xnet/xdns"
	"golang.org/x/net/dns/dnsmessage"
)

func showAnswerSection(resources []dnsmessage.Resource) {
	if !qopts.has(boolOptAnswer) || len(resources) == 0 {
		return
	}
	if qopts.has(boolOptComments) && !qopts.has(boolOptShort) {
		fmt.Println(";; ANSWER SECTION:")
	}
	for _, r := range resources {
		if !qopts.has(boolOptShort) {
			fmt.Printf("%s\t%d\t%s\t%s\t",
				r.Header.Name,
				r.Header.TTL,
				xdns.Class(r.Header.Class),
				xdns.Type(r.Header.Type),
			)
		}
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
	if qopts.has(boolOptComments) {
		fmt.Println()
	}
}

func showAuthoritySection(resources []dnsmessage.Resource) {
	if qopts.has(boolOptShort) || !qopts.has(boolOptAuthority) ||
		len(resources) == 0 {
		return
	}
	if qopts.has(boolOptComments) {
		fmt.Println(";; AUTHORITY SECTION:")
	}
	for _, r := range resources {
		fmt.Println(r.Header)
	}
	if qopts.has(boolOptComments) {
		fmt.Println()
	}
}

func showCmd() {
	if gopts.has(boolOptShort) || !gopts.has(boolOptCmd) {
		return
	}
	fmt.Printf("; <<>> goes/net-tool.DiG %s <<>> %s\n", ver, cmd)
	fmt.Println()
}

func showHeader(msg *dnsmessage.Message) {
	if qopts.has(boolOptShort) || !qopts.has(boolOptComments) {
		return
	}
	fmt.Println(";; Got answer:")
	fmt.Printf(";; ->>HEADER<<- opcode: %s, status: %s, id: %d\n",
		xdns.OpCodeName[msg.OpCode],
		xdns.RCodeName[msg.RCode],
		msg.ID,
	)
	fmt.Printf(";; flags: %s; QUERY: %d, ANSWER: %d, AUTHORITY: %d, "+
		"ADDITIONAL: %d\n",
		xdns.HeaderFlagNames(&msg.Header),
		len(msg.Questions),
		len(msg.Answers),
		len(msg.Authorities),
		len(msg.Additionals),
	)
	fmt.Println()
}

func showOptPseudoSection(resources []dnsmessage.Resource) {
	if qopts.has(boolOptShort) || !qopts.has(boolOptAdditional) ||
		len(resources) == 0 {
		return
	}
	if qopts.has(boolOptComments) {
		fmt.Println(";; OPT PSEUDOSECTION::")
	}
	for _, r := range resources {
		fmt.Print("; EDNS: version: ", xdns.EDNSVersion(r.Header.TTL))
		if xdns.HasEDNS0DNSSECOK(r.Header.TTL) {
			fmt.Print(" do")
		}
		if mbz := xdns.EDNS0MBZ(r.Header.TTL); mbz != 0 {
			fmt.Printf("; MBZ: %#.4x, udp: ", mbz)
		} else {
			fmt.Print("; udp: ")
		}
		fmt.Println(r.Header.Class)
	}
	if qopts.has(boolOptComments) {
		fmt.Println()
	}
}

func showQuestionSection(questions []dnsmessage.Question) {
	if qopts.has(boolOptShort) || !qopts.has(boolOptQuestion) ||
		len(questions) == 0 {
		return
	}
	if qopts.has(boolOptComments) {
		fmt.Println(";; QUESTION SECTION:")
	}
	for _, q := range questions {
		fmt.Printf(";%s\t%s\t%s\n",
			q.Name,
			xdns.Class(q.Class),
			xdns.Type(q.Type),
		)
	}
	if qopts.has(boolOptComments) {
		fmt.Println()
	}
}

func showStats(beg, end time.Time, n int, ra net.Addr) {
	if qopts.has(boolOptShort) || !qopts.has(boolOptStats) {
		return
	}
	fmt.Print(";; Query time: ")
	dur := end.Sub(beg)
	if flags.u {
		fmt.Println(dur.Microseconds(), "µsec")
	} else {
		fmt.Println(dur.Milliseconds(), "msec")
	}
	fmt.Printf(";; SERVER: %s#%d(%v\n", svr, flags.p, ra)
	fmt.Println(";; WHEN:", end.Format("Mon Jan 01 15:04:05 MST 2006"))
	fmt.Println(";; MSG SIZE  rcvd:", n)
}
