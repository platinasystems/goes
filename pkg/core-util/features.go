// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package core-util contain [goes.Features] that mimic Unix coreutils.
package core_util

import (
	"github.com/platinasystems/goes/v2/pkg/core-util/cat"
	"github.com/platinasystems/goes/v2/pkg/core-util/echo"
	"github.com/platinasystems/goes/v2/pkg/core-util/env"
	"github.com/platinasystems/goes/v2/pkg/core-util/hostname"
	"github.com/platinasystems/goes/v2/pkg/core-util/id"
)

var Features = map[string]any{
	"cat":      cat.Cat,
	"echo":     echo.Echo,
	"env":      env.Env,
	"hostname": hostname.Hostname,
	"id":       id.Id,
}
