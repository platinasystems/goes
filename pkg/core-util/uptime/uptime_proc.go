// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build linux

package uptime

import (
	"fmt"
	"os"
	"time"
)

func uptime() error {
	const fn = "/proc/uptime"
	var uptimeSec, uptimeMs, cpuidleSec, cpuidleMs time.Duration

	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := fmt.Fscanf(f, "%d.%d %d.%d", &uptimeSec, &uptimeMs,
		&cpuidleSec, &cpuidleMs)
	if err != nil {
		return err
	}
	if n != 4 {
		return fmt.Errorf("%s: scanned %d/4 fields", fn, n)
	}
	uptime := uptimeSec*time.Second + uptimeMs*time.Millisecond
	cpuidle := cpuidleSec*time.Second + cpuidleMs*time.Millisecond

	fmt.Printf("Uptime is %s, idle time %s\n", uptime, cpuidle)

	return nil
}
