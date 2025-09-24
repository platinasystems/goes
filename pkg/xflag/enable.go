// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xflag

import (
	"flag"
	"fmt"
	"strconv"
)

type Enable struct {
	name,
	usage string
	enable func() error
	val    bool
}

// After defined within a [flag.FlagSet] an [Enable] flag calls the given
// function when set and, if that returns nil, sets true Value.
func NewEnable(name, usage string, enable func() error) *Enable {
	return &Enable{
		name:   name,
		usage:  usage,
		enable: enable,
	}
}

func (p *Enable) Define() {
	p.DefineIn(flag.CommandLine)
}

func (p *Enable) DefineIn(fs *flag.FlagSet) {
	fs.Var(p, p.name, p.usage)
}

func (Enable) IsBoolFlag() bool {
	return true
}

func (p *Enable) String() string {
	return fmt.Sprint(p.val)
}

func (p *Enable) Set(s string) error {
	if len(s) > 0 {
		if t, err := strconv.ParseBool(s); !t || err != nil {
			return err
		}
	}
	err := p.enable()
	p.val = err == nil
	return err
}

func (p *Enable) Value() bool {
	return p.val
}
