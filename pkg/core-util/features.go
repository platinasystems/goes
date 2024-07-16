// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package core-util contain [goes.Features] that mimic Unix coreutils.
package core_util

var Features = map[string]any{
	"cat":      Cat,
	"echo":     Echo,
	"env":      Env,
	"hostname": Hostname,
	"id":       Id,
}
