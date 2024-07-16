// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This provides an approximation of the Unix net-tools command.
package main

import (
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/net-tool/netstat"
)

var features = map[string]any{
	"netstat": netstat.Feature,
}

func init() { goes.Install(features) }

func main() { goes.Main() }
