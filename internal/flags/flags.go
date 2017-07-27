// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package flags parses boolean options from command arguments.
package flags

type Flags struct {
	ByName  ByName
	aliases map[string]string
}

type ByName map[string]bool
type Aliases map[string]string

// Define and parse boolean flags from command arguments.
//
// If an argument has a leading hyphen ('-') followed by runes that all match
// '-?' flags, the respective flags are set and the argument is removed from
// the returned list, e.g.
//
//	flag, args := flags.New([]string{"-abc"}, "-a", "-b", "-c")
//
// results in
//
//	flag.ByName("-a") == true
//	flag.ByName("-b") == true
//	flag.ByName("-c") == true
//	args == []string{}
//
// whereas
//
//	flag, args := flags.New([]string{"-abcd"}, "-a", "-b", "-c")
//
// results in
//
//	flag.ByName["-a"] == false
//	flag.ByName["-b"] == false
//	flag.ByName["-c"] == false
//	args == []string{"-abcd"}
//
// Flags may be defined with strings or string slices that include aliases of
// the first entry, e.g.
//
//	flag, args := flags.New([]string{"-color"}, "-a", "-b",
//		[]string{"-c", "-color", "-colour"})
//
// results in
//
//	flag.ByName("-a") == false
//	flag.ByName("-b") == false
//	flag.ByName("-c") == true
//	args == []string{}
func New(args []string, flags ...interface{}) (*Flags, []string) {
	p := &Flags{
		ByName:  make(ByName),
		aliases: make(Aliases),
	}
	args = p.More(args, flags...)
	return p, args
}

// Define and parse more flags from command arguments.
func (p *Flags) More(args []string, flags ...interface{}) []string {
	for _, flag := range flags {
		switch t := flag.(type) {
		case string:
			p.ByName[t] = false
		case []string:
			p.ByName[t[0]] = false
			for _, aka := range t[1:] {
				p.aliases[aka] = t[0]
			}
		}
	}
	return p.Parse(args)
}

// Parse predefined flags from command arguments.
func (p *Flags) Parse(args []string) []string {
	for i := 0; i < len(args); {
		if k, found := p.aliases[args[i]]; found {
			p.ByName[k] = true
			if i < len(args)-1 {
				copy(args[i:], args[i+1:])
			}
			args = args[:len(args)-1]
		} else if _, found := p.ByName[args[i]]; found {
			p.ByName[args[i]] = true
			if i < len(args)-1 {
				copy(args[i:], args[i+1:])
			}
			args = args[:len(args)-1]
		} else if len(args[i]) > 0 && args[i][0] == '-' {
			var set []string
			for _, c := range args[i][1:] {
				s := string([]rune{'-', c})
				if _, found := p.ByName[s]; found {
					set = append(set, s)
				} else {
					set = set[:0]
					break
				}
			}
			if len(set) > 0 {
				for _, s := range set {
					p.ByName[s] = true
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
	return args
}

// Reset all flags.
func (p *Flags) Reset() {
	for k := range p.ByName {
		p.ByName[k] = false
	}
}
