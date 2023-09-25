// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package host

import (
	"os"
	"sync"
)

var Name = sync.OnceValue(hostname)

func Rename(text string) error {
	if err := rename(text); err != nil {
		return err
	}
	Name = sync.OnceValue(hostname)
	return nil
}

func hostname() string {
	s, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return s
}
