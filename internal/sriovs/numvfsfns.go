// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package sriovs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// /sys/class/net/DEV/device is a symlink to the bus id so, it's the
// best thing to sort on to have consistent interfaces
type ByBusId []string

func NumvfsFns() ([]string, error) {
	fns, err := filepath.Glob("/sys/class/net/*/device/sriov_numvfs")
	if err != nil {
		return nil, err
	}
	if len(fns) == 0 {
		return fns, fmt.Errorf("don't have an SRIOV capable device")
	}
	sort.Sort(ByBusId(fns))
	return fns, nil
}

func (fns ByBusId) Len() int {
	return len(fns)
}

func (fns ByBusId) Less(i, j int) bool {
	iln, _ := os.Readlink(filepath.Dir(fns[i]))
	jln, _ := os.Readlink(filepath.Dir(fns[j]))
	return filepath.Base(iln) < filepath.Base(jln)
}

func (fns ByBusId) Swap(i, j int) {
	fns[i], fns[j] = fns[j], fns[i]
}
