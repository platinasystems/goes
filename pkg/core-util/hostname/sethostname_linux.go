// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hostname

import "golang.org/x/sys/unix"

func Sethostname(name string) error {
	return unix.Sethostname([]byte(name))
}
