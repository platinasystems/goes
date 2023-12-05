// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package wctx

import (
	"io"
	"os"

	"github.com/platinasystems/goes/v2/pkg/context"
)

var Parameter = context.NewParameter(io.Writer(os.Stdout))
