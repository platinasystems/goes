// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bind

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnspkt"
)

//go:embed dig_options.txt
var digOptionsTxt string

const (
	digBoolOptAAOnly uint64 = 1 << iota
	digBoolOptAdditional
	digBoolOptADFlag
	digBoolOptAll
	digBoolOptAnswer
	digBoolOptAuthority
	digBoolOptBadCookie
	digBoolOptBestEffort
	digBoolOptCd
	digBoolOptClass
	digBoolOptCmd
	digBoolOptComments
	digBoolOptCookie
	digBoolOptCrypto
	digBoolOptDefName
	digBoolOptDNS64Prefix
	digBoolOptDNSSec
	digBoolOptEDNS
	digBoolOptEDNSNegotiation
	digBoolOptExpire
	digBoolOptFail
	digBoolOptHeaderOnly
	digBoolOptHTTPS
	digBoolOptHTTPSGet
	digBoolOptHTTPSPost
	digBoolOptHTTPSPlain
	digBoolOptIdentify
	digBoolOptIdn
	digBoolOptIgnore
	digBoolOptkeepAlive
	digBoolOptKeepOpen
	digBoolOptMultiline
	digBoolOptNSID
	digBoolOptNSSearch
	digBoolOptOnesoa
	digBoolOptOpCode
	digBoolOptQr
	digBoolOptQuestion
	digBoolOptRAFlag
	digBoolOptRDFlag
	digBoolOptRecurse
	digBoolOptRRComments
	digBoolOptSearch
	digBoolOptShort
	digBoolOptShowBadCookie
	digBoolOptShowSearch
	digBoolOptStats
	digBoolOptTcFlag
	digBoolOptTCP
	digBoolOptTLS
	digBoolOptTrace
	digBoolOptTtlID
	digBoolOptTtlUnits
	digBoolOptUnknownformat
	digBoolOptVc
	digBoolOptYaml
	digBoolOptZFlag
	digBoolOpts
)

var digBoolOptNamed = map[string]uint64{
	"aaonly":          digBoolOptAAOnly,
	"additional":      digBoolOptAdditional,
	"adflag":          digBoolOptADFlag,
	"all":             digBoolOptAll,
	"answer":          digBoolOptAnswer,
	"authority":       digBoolOptAuthority,
	"badcookie":       digBoolOptBadCookie,
	"besteffort":      digBoolOptBestEffort,
	"cd":              digBoolOptCd,
	"class":           digBoolOptClass,
	"cmd":             digBoolOptCmd,
	"comments":        digBoolOptComments,
	"cookie":          digBoolOptCookie,
	"crypto":          digBoolOptCrypto,
	"defname":         digBoolOptDefName,
	"dns64prefix":     digBoolOptDNS64Prefix,
	"dnssec":          digBoolOptDNSSec,
	"edns":            digBoolOptEDNS,
	"ednsnegotiation": digBoolOptEDNSNegotiation,
	"expire":          digBoolOptExpire,
	"fail":            digBoolOptFail,
	"headeronly":      digBoolOptHeaderOnly,
	"https":           digBoolOptHTTPS,
	"https-get":       digBoolOptHTTPSGet,
	"https-post":      digBoolOptHTTPSPost,
	"https-plain":     digBoolOptHTTPSPlain,
	"identify":        digBoolOptIdentify,
	"idn":             digBoolOptIdn,
	"ignore":          digBoolOptIgnore,
	"keepalive":       digBoolOptkeepAlive,
	"keepopen":        digBoolOptKeepOpen,
	"multiline":       digBoolOptMultiline,
	"nsid":            digBoolOptNSID,
	"nssearch":        digBoolOptNSSearch,
	"onesoa":          digBoolOptOnesoa,
	"opcode":          digBoolOptOpCode,
	"qr":              digBoolOptQr,
	"question":        digBoolOptQuestion,
	"raflag":          digBoolOptRAFlag,
	"rdflag":          digBoolOptRDFlag,
	"recurse":         digBoolOptRecurse,
	"rrcomments":      digBoolOptRRComments,
	"search":          digBoolOptSearch,
	"short":           digBoolOptShort,
	"showbadcookie":   digBoolOptShowBadCookie,
	"showsearch":      digBoolOptShowSearch,
	"stats":           digBoolOptStats,
	"tcflag":          digBoolOptTcFlag,
	"tcp":             digBoolOptTCP,
	"tls":             digBoolOptTLS,
	"trace":           digBoolOptTrace,
	"ttlid":           digBoolOptTtlID,
	"ttlunits":        digBoolOptTtlUnits,
	"unknownformat":   digBoolOptUnknownformat,
	"vc":              digBoolOptVc,
	"yaml":            digBoolOptYaml,
	"zflag":           digBoolOptZFlag,
}

const (
	digUintOptBufSize = iota
	digUintOptCookieValue
	digUintOptEDNSVersion
	digUintOptEDNSCode
	digUintOptFuzzTime
	digUintOptHTTPPort
	digUintOptNDots
	digUintOptOpCodeValue
	digUintOptPadding
	digUintOptQID
	digUintOptRetry
	digUintOptSplit
	digUintOptTimeout
	digUintOptTries
	digUintOpts
)

var digUintOptNamed = map[string]int{
	"bufsize":      digUintOptBufSize,
	"cookie-value": digUintOptCookieValue,
	"edns-version": digUintOptEDNSVersion,
	"edns-code":    digUintOptEDNSCode,
	"fuzz-time":    digUintOptFuzzTime,
	"http-port":    digUintOptHTTPPort,
	"ndots":        digUintOptNDots,
	"opcode-value": digUintOptOpCodeValue,
	"padding":      digUintOptPadding,
	"qid":          digUintOptQID,
	"retry":        digUintOptRetry,
	"split":        digUintOptSplit,
	"timeout":      digUintOptTimeout,
	"tries":        digUintOptTries,
}

const (
	digStringOptHTTPSQuery = iota
	digStringOptHTTPSGetQuery
	digStringOptHTTPSPosttQuery
	digStringOptSubnet
	digStringOptTLSCA
	digStringOptTLSCertFile
	digStringOptTLSHostName
	digStringOpts
)

var digStringOptNamed = map[string]int{
	"httpsquery":      digStringOptHTTPSQuery,
	"httpsgetquery":   digStringOptHTTPSGetQuery,
	"httpsposttquery": digStringOptHTTPSPosttQuery,
	"subnet":          digStringOptSubnet,
	"tlsca":           digStringOptTLSCA,
	"tlscertfile":     digStringOptTLSCertFile,
	"tlshostname":     digStringOptTLSHostName,
}

type digOptions struct {
	bools   uint64
	uints   [digUintOpts]uint
	strings [digStringOpts]string
}

var digGlobalOptions = digOptions{
	bools: digBoolOptAdditional |
		digBoolOptAnswer |
		digBoolOptAuthority |
		digBoolOptCmd |
		digBoolOptComments |
		digBoolOptCrypto |
		digBoolOptQuestion |
		digBoolOptStats,
	uints: [digUintOpts]uint{
		digUintOptBufSize: xdnspkt.Cap,
		digUintOptSplit:   56,
	},
}

func (opts *digOptions) clone() *digOptions {
	clone := new(digOptions)
	clone.bools = opts.bools
	copy(clone.uints[:], opts.uints[:])
	copy(clone.strings[:], opts.strings[:])
	return clone
}

func (opts *digOptions) has(mask uint64) bool {
	return (opts.bools & mask) != 0
}

func (opts *digOptions) parse(args []string) ([]string, error) {
	var err error
	for err == nil && len(args) > 0 && strings.HasPrefix(args[0], "+") {
		s := strings.TrimPrefix(args[0], "+")
		args = args[1:]
		if s == "noall" {
			opts.bools = 0
		} else if s == "all" {
			opts.bools = digBoolOpts - 1
		} else if strings.HasPrefix(s, "no") {
			err = opts.reset(strings.TrimPrefix(s, "no"))
		} else {
			err = opts.set(s)
		}
	}
	return args, err
}

func (opts *digOptions) reset(s string) error {
	if opt, ok := digBoolOptNamed[s]; ok {
		opts.bools &^= opt
		return nil
	}
	return xerrors.Invalid("option", s)
}

func (opts *digOptions) set(s string) error {
	if eqi := strings.Index(s, "="); eqi < 0 {
		if opt, ok := digBoolOptNamed[s]; ok {
			opts.bools |= opt
			return nil
		}
	} else {
		name, val := s[:eqi], s[eqi+1:]
		if opt, ok := digStringOptNamed[name]; ok {
			opts.strings[opt] = val
			return nil
		} else if opt, ok := digUintOptNamed[s]; ok {
			_, err := fmt.Sscan(s, &opts.uints[opt])
			if err != nil {
				err = xerrors.Label(err, "option", name)
			}
			return err
		}
	}
	return xerrors.Invalid("option", s)
}
