// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show

import (
	"bytes"
	"context"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/template"
)

const Usage = `
usage: {{.}}
	Format named value.
`

var nl = []byte("\n")

func EnsureNewLine(w io.Writer, text []byte) error {
	_, err := w.Write(text)
	if err == nil && !bytes.HasSuffix(text, nl) {
		_, err = w.Write(nl)
	}
	return err
}

type JSON struct{ json.Marshaler }

func (t JSON) Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	switch path[1] {
	case "complete":
		return nil
	case "help":
		copy(path[1:], path[2:])
		path = path[:len(path)-1]
		return template.Must(template.New("usage").Parse(Usage[1:])).
			Execute(w, strings.Join(path, " "))
	}
	if len(args) > 0 {
		return fmt.Errorf("%v unexpected", args)
	}
	data, err := t.MarshalJSON()
	if err == nil {
		_, err = w.Write(data)
	}
	return err
}

type Text struct{ encoding.TextMarshaler }

func (t Text) Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	switch path[1] {
	case "complete":
		return nil
	case "help":
		copy(path[1:], path[2:])
		path = path[:len(path)-1]
		return template.Must(template.New("usage").Parse(Usage[1:])).
			Execute(w, strings.Join(path, " "))
	}
	if len(args) > 0 {
		return fmt.Errorf("%v unexpected", args)
	}
	text, err := t.MarshalText()
	if err == nil {
		err = EnsureNewLine(w, text)
	}
	return err
}

type ContextTextMarshaler interface {
	MarshalTextContext(context.Context) ([]byte, error)
}

type TextContext struct{ ContextTextMarshaler }

func (t TextContext) Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	switch path[1] {
	case "complete":
		return nil
	case "help":
		copy(path[1:], path[2:])
		path = path[:len(path)-1]
		return template.Must(template.New("usage").Parse(Usage[1:])).
			Execute(w, strings.Join(path, " "))
	}
	if len(args) > 0 {
		return fmt.Errorf("%v unexpected", args)
	}
	text, err := t.MarshalTextContext(ctx)
	if err == nil {
		err = EnsureNewLine(w, text)
	}
	return err
}

type KeyTextMarshaler interface {
	MarshalKeyText(k string) ([]byte, error)
}

type KeyText struct{ KeyTextMarshaler }

func (kt KeyText) Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	switch path[1] {
	case "complete":
		return nil
	case "help":
		copy(path[1:], path[2:])
		path = path[:len(path)-1]
		return template.Must(template.New("usage").Parse(`
usage: {{.}} <key>
	Print the key value.
`[1:])).Execute(w, strings.Join(path, " "))
	}
	switch len(args) {
	case 0:
		return errors.New("missing <key>")
	case 1:
	default:
		return fmt.Errorf("%v unexpected", args[1:])
	}
	text, err := kt.MarshalKeyText(args[0])
	if err == nil {
		err = EnsureNewLine(w, text)
	}
	return err
}

type PathKeyText struct{ KeyTextMarshaler }

func (pkt PathKeyText) Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	switch path[1] {
	case "complete":
		return nil
	case "help":
		copy(path[1:], path[2:])
		path = path[:len(path)-1]
		return template.Must(template.New("usage").Parse(Usage[1:])).
			Execute(w, strings.Join(path, " "))
	}
	if len(args) > 0 {
		return fmt.Errorf("%v unexpected", args)
	}
	text, err := pkt.MarshalKeyText(path[len(path)-1])
	if err == nil {
		err = EnsureNewLine(w, text)
	}
	return err
}
