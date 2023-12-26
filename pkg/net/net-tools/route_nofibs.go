// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !freebsd

package net_tools

const haveFibs = false

const routeFlushUsage = `
usage: {{branch .}} [<option>]... flush
Remove all routes.`

const routeMonitorUsage = `
usage: {{branch .}} [<option>]... monitor
Continuously report route changes.`
