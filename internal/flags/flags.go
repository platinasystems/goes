// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package flags parses boolean options from command arguments.
package flags

type Flag map[string]bool

// New parses boolean flags from the given command arguments.
// If an argument has a leading hyphen ('-') followed by runes that each match
// '-?' flags, all of these flags are set and the argument is removed from the
// returned list.
func New(args []string, flags ...string) (Flag, []string) {
	flag := Flag(make(map[string]bool))
	for _, s := range flags {
		flag[s] = false
	}

	for i := 0; i < len(args); {
		if _, found := flag[args[i]]; found {
			flag[args[i]] = true
			if i < len(args)-1 {
				copy(args[i:], args[i+1:])
			}
			args = args[:len(args)-1]
		} else if len(args[i]) > 0 && args[i][0] == '-' {
			set := make([]string, 0, len(flags))
			for _, c := range args[i][1:] {
				s := string([]rune{'-', c})
				if _, found := flag[s]; found {
					set = append(set, s)
				} else {
					set = set[:0]
					break
				}
			}
			if len(set) > 0 {
				for _, s := range set {
					flag[s] = true
				}
				if i < len(args)-1 {
					copy(args[i:], args[i+1:])
				}
				args = args[:len(args)-1]
			} else {
				i++
			}
		} else {
			i++
		}
	}
	return flag, args
}

// Aka will or the named flag with each of the given aliases.
func (flag Flag) Aka(name string, aliases ...string) {
	for _, alias := range aliases {
		if !flag[name] {
			flag[name] = flag[alias]
		}
	}
}
