// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Slice a string into args while combining single, double, or backslash
// escaped spaced arguments, e.g.:
//
//	echo hello\ beautiful\ world
//	echo "hello 'beautiful world'"
//	echo 'hello \"beautiful world\"'
package fields

import (
	"regexp"
	"strings"
)

var re *regexp.Regexp

func New(s string) []string {
	if re == nil {
		re = regexp.MustCompile("'.+'|\".+\"|\\S+")
	}
	args := re.FindAllString(s, -1)
	for i, arg := range args {
		if arg[0] == '"' || arg[0] == '\'' {
			args[i] = arg[1 : len(arg)-1]
		}
	}
	for i := 0; i < len(args); {
		c := args[i][:1]
		if strings.HasSuffix(args[i], "\\") {
			args[i] = args[i][:len(args[i])-1] + " "
			if i < len(args)-1 {
				args[i] += args[i+1]
				if i < len(args)-2 {
					args = append(args[:i+1],
						args[i+2:]...)
				} else {
					args = args[:i+1]
				}
			}
		} else if c == "|" && len(args[i]) > 1 {
			args = append(args[:i], append(
				[]string{c, args[i][1:]},
				args[i+1:]...)...)
		} else if c == "<" || c == ">" {
			li := strings.LastIndex(args[i], c)
			if li >= 0 && li < len(args[i])-1 {
				arg := args[i]
				sym := arg[:li+1]
				fn := arg[li+1:]
				args = append(args[:i], append(
					[]string{sym, fn},
					args[i+1:]...)...)
				i += 2
			} else {
				i++
			}
		} else {
			i++
		}
	}
	return args
}
