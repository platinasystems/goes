// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"

	"github.com/platinasystems/go/internal/sriovs"
)

func delSriovs() error { return sriovs.Del(vfs) }

func newSriovs() error {
	if ver, err := deviceVersion(); err != nil {
		return err
	} else if ver > 0 {
		sriovs.VfName = func(port, subport uint) string {
			return fmt.Sprintf("eth-%d-%d", port+1, subport+1)
		}
	}
	return sriovs.New(vfs)
}
