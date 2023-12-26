// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build (aix && ppc64) || (zos && s390x) || ((freebsd || netbsd || openbsd) && (amd64 || arm || arm64 || riscv64)) || (linux && (amd64 || arm64 || loong64 || mips64 || mips64le || ppc64 || ppc64le || riscv64 || s390x || sparc64))

package netioctl

type TimeT int64
