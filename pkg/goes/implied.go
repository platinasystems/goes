// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
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

func ImpliedPrintOrSetObject(
	ctx context.Context,
	v any,
	args []string,
) error {
	var (
		data []byte
		err  error
	)
	if len(args) == 0 {
		w := xutf8.NewLastRuneWrapper(os.Stdout)
		if tm, ok := v.(encoding.TextMarshaler); ok {
			data, err = tm.MarshalText()
			if err != nil {
				return err
			}
			w.Write(data)
		} else if jm, ok := v.(json.Marshaler); ok {
			data, err = json.MarshalIndent(jm, "", "  ")
			if err != nil {
				return err
			}
			w.Write(data)
		} else if m, ok := v.(fmt.Stringer); ok {
			io.WriteString(w, m.String())
		} else if f, ok := v.(func() string); ok {
			io.WriteString(w, f())
		} else if b, ok := v.([]byte); ok {
			w.Write(b)
		} else {
			fmt.Fprint(w, v)
		}
		if w.LastWrittenRune() != '\n' {
			os.Stdout.WriteString("\n")
		}
		return nil
	}
	if args[0] == "-h" {
		xflag.TemplateUsage(`
usage: {{.Name}}
Print or set object with argument or stdin if that is “-”.
`)
		flag.CommandLine.Usage()
		return nil
	} else if args[0] != "-" {
		data = []byte(args[0])
	} else if data, err = io.ReadAll(os.Stdin); err != nil {
		return err
	}
	if ju, ok := v.(json.Unmarshaler); ok && json.Valid(data) {
		err = json.Unmarshal(data, ju)
	} else if tu, ok := v.(encoding.TextUnmarshaler); ok {
		err = tu.UnmarshalText(data)
	} else {
		_, err = fmt.Sscan(args[0], v)
	}
	return err
}

func ImpliedPrintResults(
	ctx context.Context,
	f func() (any, error),
	args []string,
) error {
	if len(args) > 0 && args[0] == "-h" {
		xflag.TemplateUsage(`
usage: {{.Name}}
Print results.
`)
		flag.CommandLine.Usage()
		return nil
	}
	w := xutf8.NewLastRuneWrapper(os.Stdout)
	v, err := f()
	if err != nil {
		return err
	}
	if b, ok := v.([]byte); ok {
		w.Write(b)
	} else {
		fmt.Fprint(w, v)
	}
	if w.LastWrittenRune() != '\n' {
		os.Stdout.WriteString("\n")
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
			PrintKeySet(MatchingKeys(m))
			return nil
		case Complete:
			PrintKeyLines(MatchingKeys(m))
			return nil
		case Help:
			xflag.TemplateUsage(`
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
		PrintKeyLines(MatchingKeys(m, args[0]))
		return nil
	} else if keys := MatchingKeys(m, args[0]); len(keys) > 0 {
		PrintKeySet(keys)
		return nil
	}
	return xerrors.Invalid(args[0])
}
