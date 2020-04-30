// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package dbg provides yet another stylized debug printer.

Usage:

	dbg.Style.Log(args...)
	dbg.Style.Logf(format, args...)

Where Style may be: NoOp, Plain, FileLine, or Func.

Nothing is printed with NoOp style, no args, or a nil args[0].

If args[0] is an error, both Log and Logf return that error; otherwise, these
return nil. Use this to log a returned error,

	return dbg.Style.Log(err)

Use style variables to selectively enable output,

	// PACKAGE.go
	var Err = dbg.NoOp

		...
		Err.Log(err)
		...

	// PACKAGE_test.go
	func TestMain(m *testing.M) {
		flag.Parse()
		if testing.Verbose() {
			Err = dbg.FileLine
		}
		os.Exit(m.Run())
	}
*/
package dbg

import (
	"bytes"
	"fmt"
	"go/build"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
)

// Styles: NoOp, Plain, FilLine, or Func.
type Style int

const (
	NoOp     Style = iota
	Plain          // TEXT
	FileLine       // github.com/platinasystems/dbg_test.go:22: TEXT
	Func           // github.com/platinasystems/dbg.Test() TEXT
	nStyles
)

var (
	writer atomic.Value
	cached struct {
		gopath, gopathsrc, wd struct {
			once sync.Once
			val  interface{}
		}
	}
)

// Atomic change of the os.Stdout default.
func Writer(w io.Writer) {
	writer.Store(w)
}

// Print style prefix, then args formated with fmt.Println.
func (style Style) Log(args ...interface{}) error {
	return style.log("", nil, args...)
}

// Print style prefix, then args formatted with fmt.Printf, and end with
// newline.
func (style Style) Logf(format string, args ...interface{}) error {
	return style.log(format, nil, args...)
}

// Return name of style.
func (style Style) String() string {
	if style > nStyles {
		return fmt.Sprint(int(style))
	}
	return []string{
		"NoOp",
		"Plain",
		"FileLine",
		"Func",
	}[style]
}

// The unused arg is to work-around this vet false positive,
//	call has arguments but no formatting directives
func (style Style) log(format string, _ interface{}, args ...interface{}) error {
	const skip = 2
	if len(args) == 0 || args[0] == nil {
		return nil
	}
	err, ok := args[0].(error)
	if !ok {
		err = nil
	}
	if style == NoOp {
		return err
	}
	buf := new(bytes.Buffer)
	if style > Plain {
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			fmt.Fprintf(buf, "pc[%#x] ", pc)
		} else {
			switch style {
			case FileLine:
				relfile, err := filepath.Rel(wd(), file)
				if err != nil || relfile[0] == '.' {
					relfile = relgopath(file)
				}
				fmt.Fprint(buf, relfile, ":", line, ": ")
			case Func:
				name := runtime.FuncForPC(pc).Name()
				fmt.Fprint(buf, name, "() ")
			}
		}
	}
	if len(format) > 0 {
		fmt.Fprintf(buf, format, args...)
		fmt.Fprintln(buf)
	} else {
		fmt.Fprintln(buf, args...)
	}
	w, ok := writer.Load().(io.Writer)
	if !ok || w == nil {
		w = os.Stdout
		defer os.Stdout.Sync()
	}
	w.Write(buf.Bytes())
	return err
}

func gopath() string {
	cached.gopath.once.Do(func() {
		s := os.Getenv("GOPATH")
		if len(s) == 0 {
			s = build.Default.GOPATH
		}
		cached.gopath.val = s
	})
	return cached.gopath.val.(string)
}

func gopathsrc() string {
	cached.gopathsrc.once.Do(func() {
		cached.gopathsrc.val = filepath.Join(gopath(), "src")
	})
	return cached.gopathsrc.val.(string)
}

func relgopath(path string) string {
	s, err := filepath.Rel(gopathsrc(), path)
	if err != nil {
		s = path
	}
	return s
}

func wd() string {
	cached.wd.once.Do(func() {
		s, err := os.Getwd()
		if err != nil {
			s = "."
		}
		cached.wd.val = s
	})
	return cached.wd.val.(string)
}
