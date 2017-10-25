// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package shellutils

// Pipeline is a slice of command lines. Each commands input is piped to
// the command before it, and output is piped to the next input.
type Pipeline struct {
	Cmds []Cmdline
}

func (pl *Pipeline) add(c *Cmdline) {
	if pl.Cmds == nil {
		pl.Cmds = make([]Cmdline, 0)
	}
	pl.Cmds = append(pl.Cmds, *c)
	*c = Cmdline{}
}
