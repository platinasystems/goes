// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This package provides the commands required of all goes machines.
package environ

import (
	"github.com/platinasystems/go/internal/platina-mk1-bmc/environ/fantray"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/environ/fsp"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/environ/i2cd"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/environ/nuvoton"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/environ/nxp"
	"github.com/platinasystems/go/internal/platina-mk1-bmc/environ/ti"
)

func New() []interface{} {
	return []interface{}{
		i2cd.New(),
		fantray.New(),
		fsp.New(),
		imx6.New(),
		w83795.New(),
		ucd9090.New(),
	}
}
