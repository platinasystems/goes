// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build !plugin

package fe1

import (
	"github.com/platinasystems/fe1"
	"github.com/platinasystems/go/vnet"
	fe1_platform "github.com/platinasystems/go/vnet/platforms/fe1"
)

func Packages() []map[string]string                      { return fe1.Packages }
func AddPlatform(v *vnet.Vnet, p *fe1_platform.Platform) { fe1.AddPlatform(v, p) }
func Init(v *vnet.Vnet, p *fe1_platform.Platform)        { fe1.Init(v, p) }
