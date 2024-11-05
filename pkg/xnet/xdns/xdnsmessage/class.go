// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	"strings"

	"github.com/platinasystems/goes/v2/pkg/integer"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

//go:generate stringer -type Class -trimprefix Class -linecomment
type Class uint16

const (
	Class0 Class = iota // class0

	ClassINET   // IN
	ClassCSNET  // CS
	ClassCHAOS  // CH
	ClassHESIOD // HS
)

const ClassANY Class = 255

func ClassNamed(name string) (Class, error) {
	name = strings.ToLower(name)
	if name == "any" || name == "all" {
		return ClassANY, nil
	}
	v, found := integer.Named(Class0, name, _Class_name_0,
		_Class_index_0[:]...)
	if found {
		return v, nil
	}
	return Class0, xerrors.Invalid("CLASS")
}

func (v Class) MarshalText() (b []byte, err error) {
	if v != Class0 {
		b = []byte(v.String())
	}
	return
}

func (p *Class) UnmarshalText(text []byte) (err error) {
	*p, err = ClassNamed(string(text))
	return
}
