// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"io"

	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
)

// Subscribe provider or consumer to exchange.
func Subscribe(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path selection.Path,
	args ...string,
) (err error) {
	if path.HasComplete() {
		return nil
	}
	if path.HasHelp() || selection.HasHelp(args) {
		path.Usage(w, "[[<dns>]:<port>]\n",
			"Register host or consumer with exchange."+
				"(default IPC)",
		)
		return nil
	}
	var reg string
	if len(args) > 0 {
		reg = args[0]
	}
	return tlsx.Subscribe(ctx, reg)
}
