// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !unix

package goes_util

import "github.com/platinasystems/goes/v2/pkg/xerrors"

// Pty is only available on Unix systems.
var Pty = xerrors.ErrUnavailable
