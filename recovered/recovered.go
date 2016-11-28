// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package recovered provides an interface wrapper that returns recovered
// panics as formated errors while ignoring io.EOF. It provides a similar
// function wrapper for go routines that recovers and logs any errors.
//
//	err := recovered.New(MainStringer).Main(args...)
//
//	go recovered.Go(func(args ...interface{}) {
//		...
//	}, args...)
package recovered

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/platinasystems/go/log"
)

type Recovered struct{ V }

type V interface {
	Main(...string) error
	String() string
}

func New(v V) Recovered { return Recovered{v} }

func (r Recovered) Main(args ...string) (err error) {
	defer func() {
		if err != nil && err != io.EOF {
			preface := fmt.Sprint(r.V.String(), ": ")
			if !strings.HasPrefix(err.Error(), preface) {
				err = fmt.Errorf("%s%v", preface, err)
			}
		}
	}()
	defer recovered(&err)
	err = r.V.Main(args...)
	return
}

func Go(f func(...interface{}), args ...interface{}) {
	var err error
	defer func() {
		if err != nil {
			pc := reflect.ValueOf(f).Pointer()
			fn, l := runtime.FuncForPC(pc).FileLine(pc)
			if i := strings.Index(fn, "src/"); i > 0 {
				fn = fn[i+4:]
			}
			log.Printf("daemon", "err", "%s:%d\n%v",
				fn, l, err)
		}
	}()
	defer recovered(&err)
	f(args...)
}

func recovered(p *error) {
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
		*p = errors.New(buf.String())
		buf.Reset()
	case error:
		*p = t
	case string:
		*p = errors.New(t)
	default:
		*p = fmt.Errorf("%v", t)
	}
}
