// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !linux

package tap

import (
	"context"
	"fmt"
	"io"
	"runtime"
)

const Key = "_tap"

func Daemon(context.Context, io.Reader, io.Writer, []string, ...string) error {
	return fmt.Errorf("%s can't tap", runtime.GOOS)
}
