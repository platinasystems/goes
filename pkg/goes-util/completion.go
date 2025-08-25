// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

import _ "embed"

//go:embed completion.bash
var bash string

//go:embed completion.zsh
var zsh string

var ShowCompletion = map[string]any{
	"bash": bash,
	"zsh":  zsh,
}
