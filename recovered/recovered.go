// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package recovered provides an interface wrapper that returns recovered
// panics as formated errors while ignoring io.EOF.
package recovered

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"
)

type Recovered struct{ V }

type V interface {
	Main(...string) error
	String() string
}

func New(v V) Recovered { return Recovered{v} }

func (recovered Recovered) Main(args ...string) (err error) {
	defer func() {
		r := recover()
		switch t := r.(type) {
		case nil:
		case runtime.Error:
			buf := new(bytes.Buffer)
			pc := make([]uintptr, 64)
			n := runtime.Callers(1, pc)
			fmt.Fprint(buf, t)
			start := 0
			for i := start; i < n; i++ {
				f := runtime.FuncForPC(pc[i])
				if f.Name() == "runtime.gopanic" ||
					strings.HasSuffix(f.Name(),
						"runtime.sigpanic") {
					start = i + 1
					break
				}
			}
			for i := start; i < n; i++ {
				f := runtime.FuncForPC(pc[i])
				file, line := f.FileLine(pc[i])
				i := strings.LastIndex(file, "src/")
				if i > 0 {
					file = file[i+len("src/"):]
				}
				fmt.Fprint(buf, "\n    ",
					filepath.Base(f.Name()), "()",
					"\n        ", file, ":", line)
			}
			err = errors.New(buf.String())
			buf.Reset()
		case error:
			err = t
		case string:
			err = errors.New(t)
		default:
			err = fmt.Errorf("%v", t)
		}
		preface := fmt.Sprint(recovered.V.String(), ": ")
		if err != nil {
			if err == io.EOF {
				err = nil
			} else if !strings.HasPrefix(err.Error(), preface) {
				err = fmt.Errorf("%s%v", preface, err)
			}
		}
	}()
	err = recovered.V.Main(args...)
	return
}
