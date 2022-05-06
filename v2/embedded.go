// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"embed"
)

type EmbeddedFile struct {
	FS   embed.FS
	Path string
}

func (f EmbeddedFile) Bytes() []byte {
	b, _ := f.FS.ReadFile(f.Path)
	return b
}

func (f EmbeddedFile) String() string {
	return string(f.Bytes())
}

func (f EmbeddedFile) Show(ctx context.Context, _ ...string) error {
	switch Preemption(ctx) {
	case "":
	case "help":
		Usage(ctx, "\nPrint embedded file.")
		fallthrough
	default:
		return nil
	}
	OutputOf(ctx).Print(f)
	return ctx.Err()
}
