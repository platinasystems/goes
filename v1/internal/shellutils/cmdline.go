// Copyright © 2017-2021 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package shellutils

import (
	"fmt"
	"path/filepath"
)

// Cmdline is a slice of Words which may be variable setting, a command,
// or arguments to that command. There is a seperate terminator which
// is either a pipeline operator (|) or a list operator (; & || &&).
type Cmdline struct {
	Cmds []Word
	Term Word
}

func (c *Cmdline) add(w *Word) {
	if c.Cmds == nil {
		c.Cmds = make([]Word, 0)
	}
	c.Cmds = append(c.Cmds, *w)
	*w = Word{}
}

// Slice takes a parsed command line and returns a
// map of the environment variables declared in the command,
// and a slice of the command and its arguments as strings
func (c *Cmdline) Slice(getenv func(string) string) (map[string]string, []string) {
	envmap := make(map[string]string)
	Cmdline := make([]string, 0)

	for _, w := range c.Cmds {
		s := ""
		isEnvset := false
		envsetOffset := 0
		for _, t := range w.Tokens {
			switch t.T {
			case TokenLiteral:
				s += t.V
			case TokenEnvget:
				s += getenv(t.V)
			case TokenEnvset:
				if !isEnvset {
					isEnvset = true
					envsetOffset = len(s)
				}
				s += t.V
			case TokenGlob:
				match, err := filepath.Glob(t.V)
				if match == nil || err != nil {
					s += t.V
					continue
				}
				s += match[0]
				if len(match) == 1 {
					continue
				}
				Cmdline = append(Cmdline, s)
				match = match[1:]
				if len(match) > 1 {
					Cmdline = append(Cmdline,
						match[:len(match)-1]...)
					match = match[len(match)-1:]
				}
				s = match[0]
			default:
				panic(fmt.Errorf("Unknown Token %v", t))
			}
		}
		if len(Cmdline) == 0 && isEnvset && envsetOffset != 0 {
			envmap[s[0:envsetOffset]] = s[envsetOffset+1:]
		} else {
			Cmdline = append(Cmdline, s)
		}
	}
	return envmap, Cmdline
}
