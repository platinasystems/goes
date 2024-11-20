// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

type TypeOPTResource []dnsmessage.Option

// CODE=DATA ...
// Where CODE is text representing an unsigned integer
// { [0-9]* || 0[0-8]* | 0x[0-0a-fA-F]* };
// and DATA is a string that may contain “\“ escaped characters..
func ParseOPT(tokens []string) (TypeOPTResource, error) {
	var r TypeOPTResource
	for len(tokens) > 0 {
		eq := strings.Index(tokens[0], "=")
		if eq <= 0 {
			break
		}
		var code uint16
		_, err := fmt.Sscan(tokens[0][:eq], &code)
		if err != nil {
			return r, err
		}
		s := tokens[0][eq+1:]
		if len(s) > 0 {
			if strings.IndexRune(s, '\\') >= 0 {
				if s[0] != '"' {
					s = "\"" + s + "\""
				}
				s, err = strconv.Unquote(s)
				if err != nil {
					return r, err
				}
			}
		}
		r = append(r, dnsmessage.Option{
			Code: code,
			Data: []byte(s),
		})
		tokens = tokens[1:]
	}
	return r, nil
}

func (v TypeOPTResource) construct(
	mb *dnsmessage.Builder, h dnsmessage.ResourceHeader,
) error {
	opt := dnsmessage.OPTResource{
		Options: []dnsmessage.Option(v),
	}
	return mb.OPTResource(h, opt)
}

func (v TypeOPTResource) String() string {
	switch len(v) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("code=%d, data=%#x", v[0].Code, v[0].Data)
	}
	w := new(strings.Builder)
	for _, opt := range v {
		fmt.Fprintf(w, "code=%d, data=%#x\n", opt.Code, opt.Data)
	}
	return w.String()
}
