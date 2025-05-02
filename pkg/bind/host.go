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

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsmessage"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnspkt"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

const (
	Host_4_Flag xflag.Xbool     = "4 Only use IPv4 query transport."
	Host_6_Flag xflag.Xbool     = "6 Only use IPv6 query transport."
	Host_A_Flag xflag.Xbool     = "A Like -a but omits RRSIG, NSEC, NSEC3"
	Host_C_Flag xflag.Xbool     = "C Compare SOA records on authoritative servers."
	Host_N_Flag xflag.Xuint     = "N Number of dots before root lookup is done."
	Host_R_Flag xflag.Xuint     = "R UDP retries."
	Host_T_Flag xflag.Xbool     = "T TCP mode."
	Host_U_Flag xflag.Xbool     = "U  UDP mode."
	Host_V_Flag xflag.Xbool     = "V Print version number and exit."
	Host_W_Flag xflag.Xduration = "W Reply wait time."
	Host_a_Flag xflag.Xbool     = "a Equivalent to -v -t ANY"
	Host_c_Flag BindClassFlag   = "c Query class for non-IN data"
	Host_i_Flag xflag.Xbool     = "i FIXME?"
	Host_l_Flag xflag.Xbool     = "l Using AXFR, lists all hosts in a domain."
	Host_m_Flag xflag.Xbool     = "m Memory debugging (trace|record|usage)."
	Host_p_Flag xflag.Xuint     = "p Server port."
	Host_r_Flag xflag.Xbool     = "r Disable recursive processing."
	Host_s_Flag xflag.Xbool     = "s A SERVFAIL response should stop query."
	Host_t_Flag BindTypeFlag    = "t  Query type."
	Host_v_Flag xflag.Xbool     = "v Verbose output."
	Host_w_Flag xflag.Xbool     = "w Wait forever for a reply."
)

func Host(ctx context.Context, args []string) error {
	var (
		name  string
		cflag xdnsmessage.Class
		tflag xdnsmessage.Type
	)

	xflag.TemplateUsage(`
usage: {{.Name}} [-flags] {name} [server]
Mimic BIND9's DNS lookup utility.

{{flags .}}`)

	Host_4_Flag.Define(false)
	Host_6_Flag.Define(false)
	Host_A_Flag.Define(false)
	Host_C_Flag.Define(false)
	Host_N_Flag.Define(0)
	Host_R_Flag.Define(3)
	Host_T_Flag.Define(false)
	Host_U_Flag.Define(true)
	Host_V_Flag.Define(false)
	Host_W_Flag.Define(30 * time.Second)
	Host_a_Flag.Define(false)
	Host_c_Flag.Define(&cflag, xdnsmessage.ClassINET)
	Host_i_Flag.Define(false)
	Host_l_Flag.Define(false)
	Host_m_Flag.Define(false)
	Host_p_Flag.Define(53)
	Host_r_Flag.Define(false)
	Host_s_Flag.Define(false)
	Host_t_Flag.Define(&tflag, xdnsmessage.TypeA)
	Host_v_Flag.Define(false, "d")
	Host_w_Flag.Define(false)

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

	if Host_V_Flag.Value() {
		if mm := xprogram.MainModule(); mm != nil {
			fmt.Println(mm.Version)
		} else {
			fmt.Println("(unavailable)")
		}
		return nil
	}
	if Host_v_Flag.Value() {
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

	pkt := xdnspkt.Pool.Alloc(0)
	defer xdnspkt.Pool.Free(pkt)

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

	types := []xdnsmessage.Type{tflag}
	if Host_a_Flag.Value() || Host_A_Flag.Value() {
		types[0] = xdnsmessage.TypeANY
	} else if tflag == xdnsmessage.TypeA {
		types = append(types, xdnsmessage.TypeAAAA, xdnsmessage.TypeMX)
	}
	var hf xdnsmessage.HF
	if !Host_r_Flag.Value() {
		hf |= xdnsmessage.HFRecursionDesired
	}
	for _, t := range types {
		var rsp xdnsmessage.Message
		req := xdnsmessage.Message{
			HF:     hf,
			OpCode: xdnsmessage.OpCodeQuery,
			Questions: []xdnsmessage.WireQuestion{{
				Name:  xdnsmessage.MakeUniqueString(name),
				Class: cflag,
				Type:  t,
			}},
		}
		if pkt, err = req.AppendTo(pkt[:0]); err != nil {
			return err
		}
		if pkt, err = rsvp(pkt); err != nil {
			return err
		}
		if err = rsp.UnmarshalBinary(pkt); err != nil {
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
