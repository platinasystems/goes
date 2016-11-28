// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package internal

import (
	"fmt"

	"github.com/platinasystems/go/gpio"
)

type PinValue struct {
	Name   string
	PinNum gpio.Pin
	Value  bool
}
type PinValues []PinValue

func (p *PinValue) String() string {
	kind := "IN"
	if p.PinNum&gpio.IsOutputHi != 0 {
		kind = "OUT HI"
	}
	if p.PinNum&gpio.IsOutputLo != 0 {
		kind = "OUT LO"
	}
	return fmt.Sprintf("%8s%32s: %v", kind, p.Name, p.Value)
}

// Implement sort.Interface
func (p PinValues) Len() int           { return len(p) }
func (p PinValues) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PinValues) Less(i, j int) bool { return p[i].Name < p[j].Name }
