// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package set

import (
	"context"
	"encoding"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/template"
)

type Text struct{ encoding.TextUnmarshaler }

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
		return template.Must(template.New("usage").Parse(`
usage: {{.}} <value>
	Set value.
`[1:])).Execute(w, strings.Join(path, " "))
	}
	switch len(args) {
	case 0:
		return errors.New("missing <value>")
	case 1:
	default:
		return fmt.Errorf("%v unexpected", args[1:])
	}
	return t.UnmarshalText([]byte(args[0]))
}

type KeyTextUnmarshaler interface {
	UnmarshalKeyText(k string, text []byte) error
}

type KeyText struct{ KeyTextUnmarshaler }

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
usage: {{.}} <key> <value>
	Set the key value.
`[1:])).Execute(w, strings.Join(path, " "))
	}
	switch len(args) {
	case 0:
		return errors.New("missing <key>")
	case 1:
		return errors.New("missing <value>")
	case 2:
	default:
		return fmt.Errorf("%v unexpected", args[2:])
	}
	return kt.UnmarshalKeyText(args[0], []byte(args[1]))
}

type PathKeyText struct{ KeyTextUnmarshaler }

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
		return template.Must(template.New("usage").Parse(`
usage: {{.}} <key> <value>
	Set named value.
`[1:])).Execute(w, strings.Join(path, " "))
	}
	switch len(args) {
	case 0:
		return errors.New("missing <value>")
	case 1:
	default:
		return fmt.Errorf("%v unexpected", args[1:])
	}
	k := path[len(path)-1]
	return pkt.UnmarshalKeyText(k, []byte(args[0]))
}
