// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || (aix && ppc) || ((freebsd || netbsd || openbsd) && 386) || (linux && (386 || arm || mips || mipsle || ppc))

package netioctl

type TimeT int32
