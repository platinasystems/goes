// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build freebsd

package net_tools

const haveFibs = true

const routeFlushUsage = `
usage: {{branch .}} [<option>]... flush [<fib>]...
Remove all routes in default or given FIBs.`

const routeMonitorUsage = `
usage: {{branch .}} [<option>]... monitor [<fib>]...
Continuously report route changes in default or given FIBs.`
