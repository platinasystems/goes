// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package xerrors

import "errors"

// Nil matching errors, e.g.
//
//	Suppress(err, net.ErrClosed, context.Canceled, io.EOF)
func Suppress(err error, suppressed ...error) error {
	if err != nil {
		for _, suppress := range suppressed {
			if errors.Is(err, suppress) {
				return nil
			}
		}
	}
	return err
}
