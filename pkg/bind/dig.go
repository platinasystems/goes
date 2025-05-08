// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bind

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

func Dig(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [@server] [+global] [[-flags] [name [TYPE] [CLASS] [+options]]
Mimic BIND9's DNS lookup utility.

{{flags .}}`)

	defineDigFlags()

	return digLookup(ctx, flag.CommandLine, nil, args)
}

var (
	// Dig Flags
	dig_4 = false
	dig_6 = false
	dig_O = false
	dig_T = false
	dig_b = ""
	dig_f = ""
	dig_k = ""
	dig_m = false
	dig_p = 53
	dig_u = false
	dig_v = false
	dig_y = ""
)

func defineDigFlags() {
	xflag.Define(&dig_4, "4", "Use IPv4 only.")
	xflag.Define(&dig_6, "6", "Use IPv6 only.")
	xflag.Define(&dig_O, "O", "Print plus (+) prefaced options and exit.")
	xflag.Define(&dig_T, "T", "Print types and exit.")
	xflag.Define(&dig_b, "b", `
Set the source IP address of the query. The address must be a
valid address on one of the host's network interfaces, or
"0.0.0.0" or "::". An optional port may be specified by
appending "#<port>"`[1:])
	xflag.Define(&dig_f, "f", `
Batch mode: dig reads a list of lookup requests to process from
the given file. Each line in the file should be organized in the
same way they would be presented as queries to dig using the
command-line interface.`[1:])
	xflag.Define(&dig_k, "k", `
Sign queries using TSIG using a key read from the given file.
Key files can be generated using tsig-keygen(8). When using TSIG
authentication with dig, the name server that is queried needs
to know the key and algorithm that is being used. In BIND, this
is done by providing appropriate key and server statements in
named.conf.`[1:])
	xflag.Define(&dig_m, "m", "Enable memory usage debugging.")
	xflag.Define(&dig_p, "p", `
Send the query to a non-standard port on the server, instead
of the defaut port 53. This option would be used to test a
name server that has been configured to listen for queries
on a non-standard port number.`[1:])
	xflag.Define(&dig_u, "u", `
This option indicates that print query times should be provided in microseconds
instead of milliseconds.`[1:])
	xflag.Define(&dig_v, "v", "Print the version number and exit.")
	xflag.Define(&dig_y, "y", `
Sign queries using TSIG with the given authentication key.
keyname is the name of the key, and secret is the base64 encoded
shared secret.  hmac is the name of the key algorithm; valid
choices are hmac-md5, hmac-sha1, hmac-sha224, hmac-sha256,
hmac-sha384, or hmac-sha512. If hmac is not specified, the
default is hmac-md5 or if MD5 was disabled hmac-sha256.

NOTE: You should use the -k option and avoid the -y option,
because with -y the shared secret is supplied as a command line
argument in clear text. This may be visible in the output from
ps(1) or in a history file maintained by the user's shell.`[1:])
}

func defineDigPerLookupFlags(fs *flag.FlagSet) (perlu struct {
	c       xdnsmessage.Class
	t       xdnsmessage.Type
	i, q, x string
}) {
	perlu.c = xdnsmessage.Class0
	xflag.DefineIn(fs, &perlu.c, "c", `
Set the query class. { ANY, CH, CS, HS, IN }`[1:])

	perlu.i = ""
	xflag.DefineIn(fs, &perlu.i, "i", `
Do reverse IPv6 lookups using the obsolete RFC1886 IP6.INT
domain, which is no longer in use. Obsolete bit string label
queries (RFC2874) are not attempted.`[1:])

	perlu.q = ""
	xflag.DefineIn(fs, &perlu.q, "q", `
Query the flagged name instead of positional argument.`[1:])

	perlu.t = xdnsmessage.Type0
	xflag.DefineIn(fs, &perlu.t, "t", `
The resource record type to query. It can be any valid query
type which is supported in BIND 9. The default query type is
"A", unless the -x option is supplied to indicate a reverse
lookup. A zone transfer can be requested by specifying a type of
AXFR. When an incremental zone transfer (IXFR) is required, set
the type to ixfr=N. The incremental zone transfer will contain
the changes made to the zone since the serial number in the
zone's SOA record was N.`[1:])

	perlu.x = ""
	xflag.DefineIn(fs, &perlu.x, "x", `
Simplified reverse lookups, for mapping addresses to names. The
addr is an IPv4 address in dotted-decimal notation, or a
colon-delimited IPv6 address. When the -x is used, there is no
need to provide the name, class and type arguments.  dig
automatically performs a lookup for a name like
94.2.0.192.in-addr.arpa and sets the query type and class to PTR
and IN respectively. IPv6 addresses are looked up using nibble
format under the IP6.ARPA domain (but see also the -i option).`[1:])
	return
}

func digLookup(
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
	args, err = digGlobalOptions.parse(args)
	if err != nil {
		return err
	}

	perlu := defineDigPerLookupFlags(fs)

	if err = fs.Parse(args); err != nil {
		return err
	}

	args = fs.Args()

	if fs == flag.CommandLine {
		if dig_O {
			fmt.Print(digOptionsTxt)
			return nil
		}
		if dig_T {
			fmt.Print(xdnsmessage.TypeHelpTxt)
			return nil
		}
		ver := "(unavailable)"
		if mm := xprogram.MainModule(); mm != nil {
			ver = mm.Version
		}
		if dig_v {
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
			if dig_4 {
				nw = "udp4"
			} else if dig_6 {
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
		if !digGlobalOptions.has(digBoolOptShort) &&
			digGlobalOptions.has(digBoolOptCmd) {
			fmt.Printf("; <<>> goes/pkg/bind/dig %s <<>> %s\n",
				ver, cmd)
			fmt.Println()
		}
		if len(dig_f) > 0 {
			return digBatch(ctx, rsvp, dig_f)
		}
	}
	if s := perlu.x; len(s) > 0 {
		addr, err := netip.ParseAddr(s)
		if err != nil {
			return xerrors.Label(err, "x")
		}
		name = xdnsmessage.Reverse(addr)
		perlu.t = xdnsmessage.TypePTR
		perlu.c = xdnsmessage.ClassINET
	} else {
		if len(perlu.q) > 0 {
			name = perlu.q
		} else if len(args) == 0 {
			return xerrors.Incomplete("name")
		} else {
			name = args[0]
			args = args[1:]
		}
		if perlu.t != xdnsmessage.Type0 {
		} else if len(args) == 0 {
			perlu.t = xdnsmessage.TypeA
		} else if t, err := xdnsmessage.TypeNamed(args[0]); err == nil {
			perlu.t = t
			args = args[1:]
		} else {
			perlu.t = xdnsmessage.TypeA
		}
		if perlu.c != xdnsmessage.Class0 {
		} else if len(args) == 0 {
			perlu.c = xdnsmessage.ClassINET
		} else if c, err := xdnsmessage.ClassNamed(args[0]); err == nil {
			perlu.c = c
			args = args[1:]
		} else {
			perlu.c = xdnsmessage.ClassINET
		}
	}

	qopts := digGlobalOptions.clone()
	if args, err = qopts.parse(args); err != nil {
		return err
	}

	req := xdnsmessage.Message{
		HF:     hf,
		OpCode: xdnsmessage.OpCodeQuery,
		Questions: []xdnsmessage.WireQuestion{{
			Name:  xdnsmessage.MakeUniqueString(name),
			Class: perlu.c,
			Type:  perlu.t,
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
	if !qopts.has(digBoolOptShort) && qopts.has(digBoolOptComments) {
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
	if !qopts.has(digBoolOptShort) &&
		qopts.has(digBoolOptAdditional) &&
		len(rsp.Additionals) > 0 {
		if qopts.has(digBoolOptComments) {
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
	if !qopts.has(digBoolOptShort) &&
		qopts.has(digBoolOptQuestion) &&
		len(rsp.Questions) > 0 {
		if qopts.has(digBoolOptComments) {
			fmt.Println(";; QUESTION SECTION:")
		}
		for _, q := range rsp.Questions {
			fmt.Printf(";%-31s", q.Name)
			fmt.Printf("%-8s", q.Class)
			fmt.Printf("%-s\n", q.Type)
		}
		if qopts.has(digBoolOptComments) {
			fmt.Println()
		}
	}
	if qopts.has(digBoolOptAnswer) && len(rsp.Answers) > 0 {
		if qopts.has(digBoolOptComments) &&
			!qopts.has(digBoolOptShort) {
			fmt.Println(";; ANSWER SECTION:")
		}
		for _, a := range rsp.Answers {
			if qopts.has(digBoolOptShort) {
				fmt.Print(a)
				continue
			}
			fmt.Printf("%-24s", a.Name)
			fmt.Printf("%-8d", a.Seconds())
			fmt.Printf("%-8s", a.Class)
			fmt.Printf("%-8s", a.Type())
			xdnsmessage.LineWrap(os.Stdout, a.String(), 24+8+8+8)
		}
		if qopts.has(digBoolOptComments) {
			fmt.Println()
		}
	}
	if !qopts.has(digBoolOptShort) &&
		qopts.has(digBoolOptAuthority) &&
		len(rsp.Authorities) > 0 {
		if qopts.has(digBoolOptComments) {
			fmt.Println(";; AUTHORITY SECTION:")
		}
		for _, a := range rsp.Authorities {
			if qopts.has(digBoolOptShort) {
				fmt.Print(a)
				continue
			}
			fmt.Printf("%-24s", a.Name)
			fmt.Printf("%-8d", a.Seconds())
			fmt.Printf("%-8s", a.Class)
			fmt.Printf("%-8s", a.Type())
			xdnsmessage.LineWrap(os.Stdout, a.String(), 24+8+8+8)
		}
		if qopts.has(digBoolOptComments) {
			fmt.Println()
		}
	}
	if !qopts.has(digBoolOptShort) && qopts.has(digBoolOptStats) {
		ef := end.Format("Mon Jan 01 15:04:05 MST 2006")
		fmt.Print(";; Query time: ")
		dur := end.Sub(beg)
		if dig_u {
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
		return digLookup(ctx, digNewFlagSet(), rsvp, args)
	}
	return nil
}

func digNewFlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("dig", flag.ContinueOnError)
	xflag.TemplateUsageIn(fs, "[-flags] [name [TYPE] [CLASS] [+options]]")
	return fs
}

func digBatch(
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
		fields := strings.Fields(line)
		err = digLookup(ctx, digNewFlagSet(), rsvp, fields)
	}
	return err
}
