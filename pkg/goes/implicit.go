// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"bytes"
	"context"
	"embed"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func PrintOrReadBytes(
	ctx context.Context,
	b []byte,
	args []string,
) (err error) {
	if ContextComplete(ctx) {
		return nil
	}
	if ContextHelp(ctx) {
		return Usage(ctx, `
usage: {{branch .}} [-]
Print or overwrite object with stdin.`)
	}
	if len(args) == 0 || args[0] != "-" {
		err = ContextPrintln(ctx, string(b))
	} else {
		_, err = ContextStdin(ctx).Read(b)
	}
	return
}

func PrintBytesResult(
	ctx context.Context,
	funk func() ([]byte, error),
	args []string,
) error {
	if ContextComplete(ctx) {
		return nil
	}
	if ContextHelp(ctx) {
		return Usage(ctx, `
usage: {{branch .}}
Print result.`)
	}
	data, err := funk()
	if err == nil {
		err = ContextPrintln(ctx, string(data))
	}
	return err
}

func PrintEmbedFS(ctx context.Context, efs embed.FS, args []string) error {
	if ContextComplete(ctx) {
		return nil
	}
	if len(args) == 0 || ContextHelp(ctx) {
		return Usage(ctx, `
usage: {{branch .}}
Print embedded file.`)
	}
	b, err := efs.ReadFile(args[0])
	if err == nil {
		_, err = ContextStdout(ctx).Write(b)
	}
	return err
}

type jsoner interface {
	json.Marshaler
	json.Unmarshaler
}

func PrintOrUnmarshalJSON(
	ctx context.Context,
	v jsoner,
	args []string,
) (err error) {
	var data []byte
	if ContextComplete(ctx) {
		return nil
	}
	if ContextHelp(ctx) {
		return Usage(ctx, `
usage: {{branch .}} [-]
JSON marshal object to stdout, or with "-" option, unmarshal from stdin.`)
	}
	if len(args) == 0 || args[0] != "-" {
		data, err = json.MarshalIndent(v, "", "  ")
		if err == nil {
			err = ContextPrintln(ctx, string(data))
		}
	} else {
		data, err = io.ReadAll(ContextStdin(ctx))
		if err == nil {
			err = json.Unmarshal(data, v)
		}
	}
	return
}

func PrintOrScanObject(
	ctx context.Context,
	obj any,
	args []string,
) (err error) {
	if ContextComplete(ctx) {
		return nil
	}
	if ContextHelp(ctx) {
		return Usage(ctx, `
usage: {{branch .}} [-]
Print or scan object from stdin.`)
	}
	if len(args) == 0 || args[0] != "-" {
		var sb strings.Builder
		fmt.Fprint(&sb, obj)
		err = ContextPrintln(ctx, sb.String())
	} else {
		_, err = fmt.Fscan(ContextStdin(ctx), obj)
	}
	return
}

func PrintString(ctx context.Context, s string, args []string) error {
	if ContextComplete(ctx) {
		return nil
	}
	if ContextHelp(ctx) {
		return Usage(ctx, `
usage: {{branch .}}
Print object.`)
	}
	if len(s) == 0 {
		return nil
	}
	if strings.Index(s, "{{") >= 0 {
		s = SprintTemplate(ctx, s)
	}
	return ContextPrintln(ctx, s)
}

func PrintStringer(
	ctx context.Context,
	obj fmt.Stringer,
	args []string,
) error {
	if ContextComplete(ctx) {
		return nil
	}
	if ContextHelp(ctx) {
		return Usage(ctx, `
usage: {{branch .}}
Print object.`)
	}
	return ContextPrintln(ctx, obj.String())
}

func PrintStringResult(
	ctx context.Context,
	funk func() string,
	args []string,
) error {
	if ContextComplete(ctx) {
		return nil
	}
	if ContextHelp(ctx) {
		return Usage(ctx, `
usage: {{branch .}}
Print result.`)
	}
	return ContextPrintln(ctx, funk())
}

type texter interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

func PrintOrUnmarshalText(
	ctx context.Context,
	v texter,
	args []string,
) (err error) {
	var data []byte
	if ContextComplete(ctx) {
		return nil
	}
	if ContextHelp(ctx) {
		return Usage(ctx, `
usage: {{branch .}} [-]
Plain text marshal object to stdout, or with "-" option, unmarshal from stdin.`)
	}
	if len(args) == 0 || args[0] != "-" {
		data, err = v.MarshalText()
		if err == nil {
			if data = bytes.TrimRight(data, nl); len(data) > 0 {
				w := ContextStdout(ctx)
				_, err = w.Write(data)
				w.WriteString(nl)
			}
		}
	} else {
		data, err = io.ReadAll(ContextStdin(ctx))
		if err == nil {
			err = v.UnmarshalText(data)
		}
	}
	return
}
