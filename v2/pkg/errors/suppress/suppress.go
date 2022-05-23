// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package suppress

import "errors"

// Nil matching errors, e.g.
//	err = suppress.Errors(err, net.ErrClosed, context.Canceled, io.EOF)
func Errors(err error, suppressed ...error) error {
	if err != nil {
		for _, suppress := range suppressed {
			if errors.Is(err, suppress) {
				return nil
			}
		}
	}
	return err
}
