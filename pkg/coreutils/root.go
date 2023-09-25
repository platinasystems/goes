// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package coreutils

var Root = map[string]any{
	"cat":      Cat,
	"echo":     Echo,
	"env":      Env,
	"hostname": Hostname,
	"id":       Id,
}
