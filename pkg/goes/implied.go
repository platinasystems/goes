// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"encoding"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xutf8"
)

func ImpliedPrintOrScanObject(
	ctx context.Context,
	complete bool,
	v any,
	args []string,
) error {
	var (
		data []byte
		err  error
	)
	if len(args) == 0 {
		if complete {
			// FIXME file completion
			return nil
		} else if tm, ok := v.(encoding.TextMarshaler); ok {
			data, err = tm.MarshalText()
			if err != nil {
				return err
			}
			if data = xutf8.AlineBytes(data); len(data) > 0 {
				os.Stdout.Write(data)
			}
		} else if jm, ok := v.(json.Marshaler); ok {
			data, err = json.MarshalIndent(jm, "", "  ")
			if err != nil {
				return err
			}
			if data = xutf8.AlineBytes(data); len(data) > 0 {
				os.Stdout.Write(data)
			}
		} else if b, ok := v.([]byte); ok {
			if b = xutf8.AlineBytes(b); len(b) > 0 {
				os.Stdout.Write(b)
			}
		} else if s := xutf8.AlineString(fmt.Sprint(v)); len(s) > 0 {
			os.Stdout.WriteString(s)
		}
		return nil
	} else if args[0] == "-h" {
		xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}}
Print or scan object.
`)
		flag.CommandLine.Usage()
		return nil
	} else if args[0] == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(args[0])
	}
	if err != nil {
		return err
	}
	if ju, ok := v.(json.Unmarshaler); ok && json.Valid(data) {
		err = json.Unmarshal(data, ju)
	} else if tu, ok := v.(encoding.TextUnmarshaler); ok {
		err = tu.UnmarshalText(data)
	} else {
		_, err = fmt.Sscan(string(data), v)
	}
	return err
}

func ImpliedPrintResults(
	ctx context.Context,
	f func() (any, error),
	args []string,
) error {
	if len(args) > 0 && args[0] == "-h" {
		xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}}
Print results.
`)
		flag.CommandLine.Usage()
		return nil
	}
	v, err := f()
	if err != nil {
		return err
	}
	if b, ok := v.([]byte); ok {
		if b = xutf8.AlineBytes(b); len(b) > 0 {
			os.Stdout.Write(b)
		}
	} else if s := xutf8.AlineString(fmt.Sprint(v)); len(s) > 0 {
		os.Stdout.WriteString(s)
	}
	return nil
}

func ImpliedSelect(
	ctx context.Context,
	preempt Preemption,
	m map[string]any,
	args []string,
) error {
	if len(args) == 0 {
		switch preempt {
		case None:
			return xerrors.Incomplete("feature")
		case Complete:
			PrintMatchingKeys(m)
			return nil
		case Help:
			xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} <feature> [args]
Select feature.
`)
			flag.CommandLine.Usage()
			return nil
		}
	} else if v, ok := m[args[0]]; ok {
		name := flag.CommandLine.Name() + " " + args[0]
		xflag.Rename(flag.CommandLine, name)
		return Do(ctx, preempt, v, args[1:])
	} else if len(args) == 1 && preempt == Complete {
		PrintMatchingKeys(m, args[0])
		return nil
	}
	return xerrors.Invalid(args[0])
}
