// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.
package shellutils

// List is a slice of pipelines. The pipelines were concatenated via
// unconditional execution operators (; and &) or conditional
// execution operators (|| and &&).
type List struct {
	Cmds []Pipeline
}

func (ls *List) add(pl *Pipeline) {
	if ls.Cmds == nil {
		ls.Cmds = make([]Pipeline, 0)
	}
	ls.Cmds = append(ls.Cmds, *pl)
	*pl = Pipeline{}
}
