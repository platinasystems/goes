// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package dig

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnspkt"
)

//go:embed options.txt
var optionsTxt string

const (
	boolOptAAOnly uint64 = 1 << iota
	boolOptAdditional
	boolOptADFlag
	boolOptAll
	boolOptAnswer
	boolOptAuthority
	boolOptBadCookie
	boolOptBestEffort
	boolOptCd
	boolOptClass
	boolOptCmd
	boolOptComments
	boolOptCookie
	boolOptCrypto
	boolOptDefName
	boolOptDNS64Prefix
	boolOptDNSSec
	boolOptEDNS
	boolOptEDNSNegotiation
	boolOptExpire
	boolOptFail
	boolOptHeaderOnly
	boolOptHTTPS
	boolOptHTTPSGet
	boolOptHTTPSPost
	boolOptHTTPSPlain
	boolOptIdentify
	boolOptIdn
	boolOptIgnore
	boolOptkeepAlive
	boolOptKeepOpen
	boolOptMultiline
	boolOptNSID
	boolOptNSSearch
	boolOptOnesoa
	boolOptOpCode
	boolOptQr
	boolOptQuestion
	boolOptRAFlag
	boolOptRDFlag
	boolOptRecurse
	boolOptRRComments
	boolOptSearch
	boolOptShort
	boolOptShowBadCookie
	boolOptShowSearch
	boolOptStats
	boolOptTcFlag
	boolOptTCP
	boolOptTLS
	boolOptTrace
	boolOptTtlID
	boolOptTtlUnits
	boolOptUnknownformat
	boolOptVc
	boolOptYaml
	boolOptZFlag
	boolOpts
)

var boolOptNamed = map[string]uint64{
	"aaonly":          boolOptAAOnly,
	"additional":      boolOptAdditional,
	"adflag":          boolOptADFlag,
	"all":             boolOptAll,
	"answer":          boolOptAnswer,
	"authority":       boolOptAuthority,
	"badcookie":       boolOptBadCookie,
	"besteffort":      boolOptBestEffort,
	"cd":              boolOptCd,
	"class":           boolOptClass,
	"cmd":             boolOptCmd,
	"comments":        boolOptComments,
	"cookie":          boolOptCookie,
	"crypto":          boolOptCrypto,
	"defname":         boolOptDefName,
	"dns64prefix":     boolOptDNS64Prefix,
	"dnssec":          boolOptDNSSec,
	"edns":            boolOptEDNS,
	"ednsnegotiation": boolOptEDNSNegotiation,
	"expire":          boolOptExpire,
	"fail":            boolOptFail,
	"headeronly":      boolOptHeaderOnly,
	"https":           boolOptHTTPS,
	"https-get":       boolOptHTTPSGet,
	"https-post":      boolOptHTTPSPost,
	"https-plain":     boolOptHTTPSPlain,
	"identify":        boolOptIdentify,
	"idn":             boolOptIdn,
	"ignore":          boolOptIgnore,
	"keepalive":       boolOptkeepAlive,
	"keepopen":        boolOptKeepOpen,
	"multiline":       boolOptMultiline,
	"nsid":            boolOptNSID,
	"nssearch":        boolOptNSSearch,
	"onesoa":          boolOptOnesoa,
	"opcode":          boolOptOpCode,
	"qr":              boolOptQr,
	"question":        boolOptQuestion,
	"raflag":          boolOptRAFlag,
	"rdflag":          boolOptRDFlag,
	"recurse":         boolOptRecurse,
	"rrcomments":      boolOptRRComments,
	"search":          boolOptSearch,
	"short":           boolOptShort,
	"showbadcookie":   boolOptShowBadCookie,
	"showsearch":      boolOptShowSearch,
	"stats":           boolOptStats,
	"tcflag":          boolOptTcFlag,
	"tcp":             boolOptTCP,
	"tls":             boolOptTLS,
	"trace":           boolOptTrace,
	"ttlid":           boolOptTtlID,
	"ttlunits":        boolOptTtlUnits,
	"unknownformat":   boolOptUnknownformat,
	"vc":              boolOptVc,
	"yaml":            boolOptYaml,
	"zflag":           boolOptZFlag,
}

const (
	uintOptBufSize = iota
	uintOptCookieValue
	uintOptEDNSVersion
	uintOptEDNSCode
	uintOptFuzzTime
	uintOptHTTPPort
	uintOptNDots
	uintOptOpCodeValue
	uintOptPadding
	uintOptQID
	uintOptRetry
	uintOptSplit
	uintOptTimeout
	uintOptTries
	uintOpts
)

var uintOptNamed = map[string]int{
	"bufsize":      uintOptBufSize,
	"cookie-value": uintOptCookieValue,
	"edns-version": uintOptEDNSVersion,
	"edns-code":    uintOptEDNSCode,
	"fuzz-time":    uintOptFuzzTime,
	"http-port":    uintOptHTTPPort,
	"ndots":        uintOptNDots,
	"opcode-value": uintOptOpCodeValue,
	"padding":      uintOptPadding,
	"qid":          uintOptQID,
	"retry":        uintOptRetry,
	"split":        uintOptSplit,
	"timeout":      uintOptTimeout,
	"tries":        uintOptTries,
}

const (
	stringOptHTTPSQuery = iota
	stringOptHTTPSGetQuery
	stringOptHTTPSPosttQuery
	stringOptSubnet
	stringOptTLSCA
	stringOptTLSCertFile
	stringOptTLSHostName
	stringOpts
)

var stringOptNamed = map[string]int{
	"httpsquery":      stringOptHTTPSQuery,
	"httpsgetquery":   stringOptHTTPSGetQuery,
	"httpsposttquery": stringOptHTTPSPosttQuery,
	"subnet":          stringOptSubnet,
	"tlsca":           stringOptTLSCA,
	"tlscertfile":     stringOptTLSCertFile,
	"tlshostname":     stringOptTLSHostName,
}

type options struct {
	bools   uint64
	uints   [uintOpts]uint
	strings [stringOpts]string
}

var gopts = options{ // Global Options
	bools: boolOptAdditional |
		boolOptAnswer |
		boolOptAuthority |
		boolOptCmd |
		boolOptComments |
		boolOptCrypto |
		boolOptQuestion |
		boolOptStats,
	uints: [uintOpts]uint{
		uintOptBufSize: xdnspkt.Cap,
		uintOptSplit:   56,
	},
}

func (opts *options) clone() *options {
	clone := new(options)
	clone.bools = opts.bools
	copy(clone.uints[:], opts.uints[:])
	copy(clone.strings[:], opts.strings[:])
	return clone
}

func (opts *options) has(mask uint64) bool {
	return (opts.bools & mask) != 0
}

func (opts *options) parse(args []string) ([]string, error) {
	var err error
	for err == nil && len(args) > 0 && strings.HasPrefix(args[0], "+") {
		s := strings.TrimPrefix(args[0], "+")
		args = args[1:]
		if s == "noall" {
			opts.bools = 0
		} else if s == "all" {
			opts.bools = boolOpts - 1
		} else if strings.HasPrefix(s, "no") {
			err = opts.reset(strings.TrimPrefix(s, "no"))
		} else {
			err = opts.set(s)
		}
	}
	return args, err
}

func (opts *options) reset(s string) error {
	if opt, ok := boolOptNamed[s]; ok {
		opts.bools &^= opt
		return nil
	}
	return xerrors.Invalid("option", s)
}

func (opts *options) set(s string) error {
	if eqi := strings.Index(s, "="); eqi < 0 {
		if opt, ok := boolOptNamed[s]; ok {
			opts.bools |= opt
			return nil
		}
	} else {
		name, val := s[:eqi], s[eqi+1:]
		if opt, ok := stringOptNamed[name]; ok {
			opts.strings[opt] = val
			return nil
		} else if opt, ok := uintOptNamed[s]; ok {
			_, err := fmt.Sscan(s, &opts.uints[opt])
			if err != nil {
				err = xerrors.Label(err, "option", name)
			}
			return err
		}
	}
	return xerrors.Invalid("option", s)
}
