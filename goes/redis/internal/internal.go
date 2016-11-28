// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package internal

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/platinasystems/go/redis"
)

var these struct {
	keys  []string
	hkeys map[string][]string
}

func Fprintln(w io.Writer, s string) {
	s = Quotes(s)
	if len(s) == 0 {
		return
	}
	w.Write([]byte(s))
	if s[len(s)-1] != '\n' {
		w.Write([]byte{'\n'})
	}
}

func Quotes(s string) string {
	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return fmt.Sprintf("%q", s)
		}
	}
	return s
}

// Complete redis Key and Subkey. This skips over leading '-' prefaced flags.
func Complete(args ...string) (c []string) {
	if len(args) != 0 && strings.HasPrefix(args[0], "-") {
		args = args[1:]
	}
	switch len(args) {
	case 0:
		c, _ = redis.Keys(".*")
	case 1:
		c, _ = redis.Keys(args[0] + ".*")
	case 2:
		subkeys, _ := redis.Hkeys(args[0])
		for _, subkey := range subkeys {
			if strings.HasPrefix(subkey, args[1]) {
				c = append(c, subkey)
			}
		}
	}
	return
}
