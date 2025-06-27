// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package xerrors

import (
	"errors"
	"fmt"
	"runtime"
)

// Record a recovered panic as an error wrapped with the panic'd file name and
// line number, e.g.
//
//	func() (err error) {
//		defer Recovery(&err)
//		...
//		panic("oops")
//	}
func Recovery(p *error, suppressed ...error) {
	r := recover()
	if r == nil {
		return
	}
	err, is_error := r.(error)
	if !is_error {
		err = errors.New(fmt.Sprint(r))
	} else if err = Suppress(err, suppressed...); err == nil {
		return
	}
	if _, is_mark := err.(MarkError); is_mark {
		*p = err
		return
	}
	if *p != nil {
		err = Label(*p, err)
	}
	depth := 2
	if _, is_runtime := err.(runtime.Error); is_runtime {
		depth = 3
	}
	if _, f, l, ok := runtime.Caller(depth); ok {
		err = Label(err, FileNameMutation(f), l)
	}
	*p = err
}
