// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package dig

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

type options struct {
	aaonly          bool
	additional      bool
	adflag          bool
	all             bool
	answer          bool
	authority       bool
	badcookie       bool
	besteffort      bool
	bufsize         uint16
	cd              bool
	class           bool
	cmd             bool
	co              bool
	comments        bool
	cookie          string
	crypto          bool
	defname         bool
	dns64prefix     bool
	dnssec          bool
	domain          string
	edns            uint8
	ednsflags       uint8
	ednsnegotiation bool
	ednsopt         string
	expire          bool
	fail            bool
	fuzztime        time.Time
	headerOnly      bool
	httpPlain       string
	httpPlainGet    string
	https           string
	httpsGet        string
	httpsSkipVerify bool
	identify        bool
	idn             bool
	ignore          bool
	keepalive       bool
	keepopen        bool
	multiline       bool
	ndots           uint8
	nsid            bool
	nssearch        bool
	onesoa          bool
	opcode          uint8
	padding         uint16
	proxy           string
	proxyPlain      string
	qid             uint16
	qr              bool
	question        bool
	raflag          bool
	recurse         bool
	retry           uint8
	rrcomments      bool
	search          bool
	short           bool
	showbadcookie   bool
	showbadvers     bool
	showsearch      bool
	split           uint8
	stats           bool
	subnet          string
	tcflag          bool
	tcp             bool
	timeout         uint8
	tls             bool
	tlsCA           string
	tlsCertFile     string
	tlsHostname     string
	tlsKeyFile      string
	trace           bool
	tries           uint8
	ttlid           bool
	ttlunits        bool
	unknownformat   bool
	vc              bool
	yaml            bool
	zflag           bool
}

func (o *options) Format(w fmt.State, verb rune) {
	if o.aaonly {
		fmt.Fprint(w, " +aaonly")
	}
	if o.additional {
		fmt.Fprint(w, " +additional")
	}
	if o.adflag {
		fmt.Fprint(w, " +adflag")
	}
	if o.all {
		fmt.Fprint(w, " +all")
	}
	if o.answer {
		fmt.Fprint(w, " +answer")
	}
	if o.authority {
		fmt.Fprint(w, " +authority")
	}
	if o.badcookie {
		fmt.Fprint(w, " +badcookie")
	}
	if o.besteffort {
		fmt.Fprint(w, " +besteffort")
	}
	if o.bufsize != 0 {
		fmt.Fprintf(w, " +bufsize:%d", o.bufsize)
	}
	if o.cd {
		fmt.Fprint(w, " +cd")
	}
	if o.class {
		fmt.Fprint(w, " +class")
	}
	if o.cmd {
		fmt.Fprint(w, " +cmd")
	}
	if o.co {
		fmt.Fprint(w, " +co")
	}
	if o.comments {
		fmt.Fprint(w, " +comments")
	}
	if len(o.cookie) != 0 {
		fmt.Fprintf(w, " +cookie:%s", o.cookie)
	}
	if o.crypto {
		fmt.Fprint(w, " +crypto")
	}
	if o.defname {
		fmt.Fprint(w, " +defname")
	}
	if o.dns64prefix {
		fmt.Fprint(w, " +dns64prefix")
	}
	if o.dnssec {
		fmt.Fprint(w, " +dnssec")
	}
	if len(o.domain) != 0 {
		fmt.Fprintf(w, " +domain:%s", o.domain)
	}
	if o.edns != 0 {
		fmt.Fprintf(w, " +edns:%d", o.edns)
	}
	if o.ednsflags != 0 {
		fmt.Fprintf(w, " +ednsflags:%d", o.ednsflags)
	}
	if o.ednsnegotiation {
		fmt.Fprint(w, " +ednsnegotiation")
	}
	if len(o.ednsopt) != 0 {
		fmt.Fprintf(w, " +ednsopt:%s", o.ednsopt)
	}
	if o.expire {
		fmt.Fprint(w, " +expire")
	}
	if o.fail {
		fmt.Fprint(w, " +fail")
	}
	// FIXME fuzztime        time.Time
	if o.headerOnly {
		fmt.Fprint(w, " +header-only")
	}
	if len(o.httpPlain) != 0 {
		fmt.Fprintf(w, " +http-plain:%s", o.httpPlain)
	}
	if len(o.httpPlainGet) != 0 {
		fmt.Fprintf(w, " +http-plain-get:%s", o.httpPlainGet)
	}
	if len(o.https) != 0 {
		fmt.Fprintf(w, " +https:%s", o.https)
	}
	if len(o.httpsGet) != 0 {
		fmt.Fprintf(w, " +https-get:%s", o.httpsGet)
	}
	if o.httpsSkipVerify {
		fmt.Fprint(w, " +https-skip-verify")
	}
	if o.identify {
		fmt.Fprint(w, " +identify")
	}
	if o.idn {
		fmt.Fprint(w, " +idn")
	}
	if o.ignore {
		fmt.Fprint(w, " +ignore")
	}
	if o.keepalive {
		fmt.Fprint(w, " +keepalive")
	}
	if o.keepopen {
		fmt.Fprint(w, " +keepopen")
	}
	if o.multiline {
		fmt.Fprint(w, " +multiline")
	}
	if o.ndots != DefaultNdots {
		fmt.Fprintf(w, " +ndots:%d", o.ndots)
	}
	if o.nsid {
		fmt.Fprint(w, " +nsid")
	}
	if o.nssearch {
		fmt.Fprint(w, " +nssearch")
	}
	if o.onesoa {
		fmt.Fprint(w, " +onesoa")
	}
	if o.opcode != 0 {
		fmt.Fprintf(w, " +opcode:%d", o.opcode)
	}
	if o.padding != 0 {
		fmt.Fprintf(w, " +padding:%d", o.padding)
	}
	if len(o.proxy) != 0 {
		fmt.Fprintf(w, " +proxy:%s", o.proxy)
	}
	if len(o.proxyPlain) != 0 {
		fmt.Fprintf(w, " +proxy-plain:%s", o.proxyPlain)
	}
	if o.qid != 0 {
		fmt.Fprintf(w, " +qid:%d", o.qid)
	}
	if o.qr {
		fmt.Fprint(w, " +qr")
	}
	if o.question {
		fmt.Fprint(w, " +question")
	}
	if o.raflag {
		fmt.Fprint(w, " +raflag")
	}
	if o.recurse {
		fmt.Fprint(w, " +recurse")
	}
	if o.retry != 0 {
		fmt.Fprintf(w, " +retry:%d", o.retry)
	}
	if o.rrcomments {
		fmt.Fprint(w, " +rrcomments")
	}
	if o.search {
		fmt.Fprint(w, " +search")
	}
	if o.short {
		fmt.Fprint(w, " +short")
	}
	if o.showbadcookie {
		fmt.Fprint(w, " +showbadcookie")
	}
	if o.showbadvers {
		fmt.Fprint(w, " +showbadvers")
	}
	if o.showsearch {
		fmt.Fprint(w, " +showsearch")
	}
	if o.split != 0 {
		fmt.Fprintf(w, " +split:%d", o.split)
	}
	if o.stats {
		fmt.Fprint(w, " +stats")
	}
	if len(o.subnet) != 0 {
		fmt.Fprintf(w, " +subnet:%s", o.subnet)
	}
	if o.tcflag {
		fmt.Fprint(w, " +tcflag")
	}
	if o.tcp {
		fmt.Fprint(w, " +tcp")
	}
	if o.timeout != 0 {
		fmt.Fprintf(w, " +timeout:%d", o.timeout)
	}
	if o.tls {
		fmt.Fprint(w, " +tls")
	}
	if len(o.tlsCA) != 0 {
		fmt.Fprintf(w, " +tls-ca:%s", o.tlsCA)
	}
	if len(o.tlsCertFile) != 0 {
		fmt.Fprintf(w, " +tls-certfile:%s", o.tlsCertFile)
	}
	if len(o.tlsHostname) != 0 {
		fmt.Fprintf(w, " +tls-hostname:%s", o.tlsHostname)
	}
	if len(o.tlsKeyFile) != 0 {
		fmt.Fprintf(w, " +tls-keyfile:%s", o.tlsKeyFile)
	}
	if o.trace {
		fmt.Fprint(w, " +trace")
	}
	if o.tries != 0 {
		fmt.Fprintf(w, " +tries:%d", o.tries)
	}
	if o.ttlid {
		fmt.Fprint(w, " +ttlid")
	}
	if o.ttlunits {
		fmt.Fprint(w, " +ttlunits")
	}
	if o.unknownformat {
		fmt.Fprint(w, " +unknownformat")
	}
	if o.vc {
		fmt.Fprint(w, " +vc")
	}
	if o.yaml {
		fmt.Fprint(w, " +yaml")
	}
	if o.zflag {
		fmt.Fprint(w, " +zflag")
	}
}

func (o *options) reset(name string) error {
	var err error
	switch name {
	case "aaonly", "aaflag":
		o.aaonly = false
	case "additional":
		o.additional = false
	case "adflag":
		o.adflag = false
	case "all":
		o.all = false
	case "answer":
		o.answer = false
	case "authority":
		o.authority = false
	case "badcookie":
		o.badcookie = false
	case "besteffort":
		o.besteffort = false
	case "bufsize":
		o.bufsize = 0
	case "cd", "cdflag":
		o.cd = false
	case "class":
		o.class = false
	case "cmd":
		o.cmd = false
	case "co", "coflag":
		o.co = false
	case "comments":
		o.comments = false
	case "cookie":
		o.cookie = ""
	case "crypto":
		o.crypto = false
	case "defname":
		o.defname = false
	case "dns64prefix":
		o.dns64prefix = false
	case "dnssec", "do":
		o.dnssec = false
	case "domain":
		o.domain = ""
	case "edns":
		o.edns = 0
	case "ednsflags":
		o.ednsflags = 0
	case "ednsnegotiation":
		o.ednsnegotiation = false
	case "ednsopt":
		o.ednsopt = ""
	case "expire":
		o.expire = false
	case "fail":
		o.fail = false
	case "fuzztime":
		o.fuzztime = time.Now()
	case "headerOnly":
		o.headerOnly = false
	case "https", "https-post":
		o.https = ""
	case "https-get":
		o.httpsGet = ""
	case "http-plain", "http-plain-post":
		o.httpPlain = ""
	case "http-plain-get":
		o.httpPlainGet = ""
	case "https-skip-verify":
		o.httpsSkipVerify = false
	case "identify":
		o.identify = false
	case "idn":
		o.idn = false
	case "ignore":
		o.ignore = false
	case "keepalive":
		o.keepalive = false
	case "keepopen":
		o.keepopen = false
	case "multiline":
		o.multiline = false
	case "ndots":
		o.ndots = 0
	case "nsid":
		o.nsid = false
	case "nssearch":
		o.nssearch = false
	case "onesoa":
		o.onesoa = false
	case "opcode":
		o.opcode = 0
	case "padding":
		o.padding = 0
	case "proxy":
		o.proxy = ""
	case "proxy-plain":
		o.proxyPlain = ""
	case "qid":
		o.qid = 0
	case "qr":
		o.qr = false
	case "question":
		o.question = false
	case "raflag":
		o.raflag = false
	case "recurse", "rdflag":
		o.recurse = false
	case "retry":
		o.retry = 0
	case "rrcomments":
		o.rrcomments = false
	case "search":
		o.search = false
	case "short":
		o.short = false
	case "showbadcookie":
		o.showbadcookie = false
	case "showbadvers":
		o.showbadvers = false
	case "showsearch":
		o.showsearch = false
	case "split":
		o.split = 0
	case "stats":
		o.stats = false
	case "subnet":
		o.subnet = ""
	case "tcflag":
		o.tcflag = false
	case "tcp":
		o.tcp = false
	case "timeout":
		o.timeout = 0
	case "tls":
		o.tls = false
	case "tls-ca":
		o.tlsCA = ""
	case "tls-certfile":
		o.tlsCertFile = ""
	case "tls-hostname":
		o.tlsHostname = ""
	case "tls-keyfile":
		o.tlsKeyFile = ""
	case "trace":
		o.trace = false
	case "tries":
		o.tries = 0
	case "ttlid":
		o.ttlid = false
	case "ttlunits":
		o.ttlunits = false
	case "unknownformat":
		o.unknownformat = false
	case "vc":
		o.vc = false
	case "yaml":
		o.yaml = false
	case "zflag":
		o.zflag = false
	default:
		err = xerrors.Invalid(name)
	}
	return err
}

func (o *options) set(name string) error {
	var err error
	var val string
	if i := strings.Index(name, "="); i > 0 {
		val = name[i+1:]
		name = name[:i]
	}
	switch name {
	case "aaonly":
		o.aaonly = true
	case "additional":
		o.additional = true
	case "adflag":
		o.adflag = true
	case "all":
		o.all = true
	case "answer":
		o.answer = true
	case "authority":
		o.authority = true
	case "badcookie":
		o.badcookie = true
	case "besteffort":
		o.besteffort = true
	case "bufsize":
		err = setUint(&o.bufsize, name, val)
	case "cd", "cdflag":
		o.cd = true
	case "class":
		o.class = true
	case "cmd":
		o.cmd = true
	case "co", "coflag":
		o.co = true
	case "comments":
		o.comments = true
	case "cookie":
		o.cookie = val
	case "crypto":
		o.crypto = true
	case "defname":
		o.defname = true
	case "dns64prefix":
		o.dns64prefix = true
	case "dnssec", "do":
		o.dnssec = true
	case "domain":
		o.domain = val
	case "edns":
		err = setUint(&o.edns, name, val)
	case "ednsflags":
		err = setUint(&o.ednsflags, name, val)
	case "ednsnegotiation":
		o.ednsnegotiation = true
	case "ednsopt":
		o.ednsopt = val
	case "expire":
		o.expire = true
	case "fail":
		o.fail = true
	case "fuzztime":
		// Fri 11 Mar 2022 04:15:29 UTC
		us := int64(1646972129)
		if len(val) > 0 {
			_, err = fmt.Sscan(val, &us)
		}
		if err == nil {
			o.fuzztime = time.Unix(us, 0)
		}
	case "header-only":
		o.headerOnly = true
	case "https", "https-post":
		if o.https = val; len(o.https) == 0 {
			o.https = "/dns-query"
		} else if !strings.HasPrefix(o.https, "/") {
			o.https = fmt.Sprint("/", o.https)
		}
	case "https-get":
		err = xerrors.Unsupported(name)
		if o.httpsGet = val; len(o.httpsGet) == 0 {
			o.httpsGet = "/dns-query"
		}
	case "http-plain", "http-plain-post":
		o.httpPlain = val
	case "http-plain-get":
		err = xerrors.Unsupported(name)
		if o.httpPlainGet = val; len(o.httpPlainGet) == 0 {
			o.httpPlainGet = "/dns-query"
		}
	case "https-skip-verify":
		o.httpsSkipVerify = true
	case "identify":
		o.identify = true
	case "idn":
		o.idn = true
	case "ignore":
		o.ignore = true
	case "keepalive":
		o.keepalive = true
	case "keepopen":
		o.keepopen = true
	case "multiline":
		o.multiline = true
	case "ndots":
		err = setUint(&o.ndots, name, val)
	case "nsid":
		o.nsid = true
	case "nssearch":
		o.nssearch = true
		o.recurse = false
	case "onesoa":
		o.onesoa = true
	case "opcode":
		err = setUint(&o.opcode, name, val)
	case "padding":
		err = setUint(&o.padding, name, val)
	case "proxy":
		o.proxy = val
	case "proxy-plain":
		o.proxyPlain = val
	case "qid":
		err = setUint(&o.qid, name, val)
	case "qr":
		o.qr = true
	case "question":
		o.question = true
	case "raflag":
		o.raflag = true
	case "recurse", "rdflag":
		o.recurse = true
	case "retry":
		err = setUint(&o.retry, name, val)
	case "rrcomments":
		o.rrcomments = true
	case "search":
		o.search = true
	case "short":
		o.short = true
	case "showbadcookie":
		o.showbadcookie = true
	case "showbadvers":
		o.showbadvers = true
	case "showsearch":
		o.showsearch = true
	case "split":
		err = setUint(&o.split, name, val)
	case "stats":
		o.stats = true
	case "subnet":
		o.subnet = val
	case "tcflag":
		o.tcflag = true
	case "tcp":
		err = xerrors.Unsupported(name)
		o.tcp = true
	case "timeout":
		err = setUint(&o.timeout, name, val)
	case "tls":
		o.tls = true
	case "tls-ca":
		o.tlsCA = val
	case "tls-certfile":
		o.tlsCertFile = val
	case "tls-hostname":
		o.tlsHostname = val
	case "tls-keyfile":
		o.tlsKeyFile = val
	case "trace":
		o.trace = true
		o.recurse = false
	case "tries":
		err = setUint(&o.tries, name, val)
	case "ttlid":
		o.ttlid = true
	case "ttlunits":
		o.ttlunits = true
	case "unknownformat":
		o.unknownformat = true
	case "vc":
		o.vc = true
	case "yaml":
		o.yaml = true
	case "zflag":
		o.zflag = true
	default:
		err = xerrors.Invalid(name)
	}
	return err
}

// If “s” has < options["ndots"],
// append options["domain"], if present,
// or resolve.search[0]
func (o *options) expandedName(s string) string {
	ndots := o.ndots
	if ndots == 0 {
		if v, ok := r.Options["ndots"]; ok {
			if u, ok := v.(uint); ok {
				ndots = uint8(u)
			} else {
				return s
			}
		} else {
			return s
		}
	} else {
		return s
	}
	if uint8(strings.Count(s, ".")) < ndots {
		if len(o.domain) > 0 {
			if !strings.HasPrefix(o.domain, ".") {
				s += "."
			}
			s += o.domain
		} else if o.search {
			x := r.Search[0]
			if !strings.HasPrefix(x, ".") {
				s += "."
			}
			s += x
		}
		if !strings.HasSuffix(s, ".") {
			s += "."
		}
	}
	return s
}

func setUint[T uint8 | uint16](p *T, name, val string) error {
	if len(val) == 0 {
		return xerrors.Incomplete(name)
	}
	_, err := fmt.Sscan(val, p)
	return err
}
