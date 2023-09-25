// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package embedded

import "embed"

type File struct {
	FS   embed.FS
	Name string
}

func (f File) Bytes() []byte {
	b, _ := f.FS.ReadFile(f.Name)
	return b
}

func (f File) String() string {
	return string(f.Bytes())
}
