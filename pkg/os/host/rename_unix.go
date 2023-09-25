// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build aix || solaris

package host

import "golang.org/x/sys/unix"

var rename = unix.Sethostname
