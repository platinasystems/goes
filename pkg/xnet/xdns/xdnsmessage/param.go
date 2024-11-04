// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import "fmt"

type Param struct{ Key, Value fmt.Stringer }

func (param Param) String() string {
	return fmt.Sprint(param.Key, "=", param.Value)
}
