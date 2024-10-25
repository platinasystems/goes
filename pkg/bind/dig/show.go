// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package dig

import (
	"fmt"
	"net"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
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
			fmt.Printf("%-24s", r.Header.Name)
			fmt.Printf("%-8d", r.Header.TTL)
			fmt.Printf("%-8s", xdnsmessage.Class(r.Header.Class))
			fmt.Printf("%-8s", xdnsmessage.Type(r.Header.Type))
		}
		fmt.Println(xdnsmessage.AnswerString(r))
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
	fmt.Printf("; <<>> goes/pkg/bind/dig %s <<>> %s\n", ver, cmd)
	fmt.Println()
}

func showHeader(msg *dnsmessage.Message) {
	if qopts.has(boolOptShort) || !qopts.has(boolOptComments) {
		return
	}
	fmt.Println(";; Got answer:")
	fmt.Printf(";; ->>HEADER<<- opcode: %s, status: %s, id: %d\n",
		xdnsmessage.OpCodeName[msg.OpCode],
		xdnsmessage.RCodeName[msg.RCode],
		msg.ID,
	)
	fmt.Printf(";; flags: %s; QUERY: %d, ANSWER: %d, AUTHORITY: %d, "+
		"ADDITIONAL: %d\n",
		xdnsmessage.HeaderFlags{msg.Header},
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
		fmt.Println(";; OPT PSEUDOSECTION:")
	}
	for _, r := range resources {
		fmt.Print("; EDNS: version: ",
			xdnsmessage.EDNSVersion(r.Header.TTL))
		if xdnsmessage.HasEDNS0DNSSECOK(r.Header.TTL) {
			fmt.Print(" do")
		}
		if mbz := xdnsmessage.EDNS0MBZ(r.Header.TTL); mbz != 0 {
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
		fmt.Printf(";%-31s", q.Name)
		fmt.Printf("%-8s", xdnsmessage.Class(q.Class))
		fmt.Printf("%-s\n", xdnsmessage.Type(q.Type))
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
	fmt.Println(";; MSG SIZE:", n)
}
