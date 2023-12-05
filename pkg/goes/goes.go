// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"bytes"
	"context"
	"embed"
	"encoding"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/context/pathctx"
	"github.com/platinasystems/goes/v2/pkg/context/rctx"
	"github.com/platinasystems/goes/v2/pkg/context/selctx"
	"github.com/platinasystems/goes/v2/pkg/context/wctx"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/log/oslog"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/termination"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

var Append, Complete, Help, Json, Tee bool
var Input, Output string
var Mode uint = 0666
var Timeout time.Duration

// Execute subsystem in an interruptible context.
func Exec(subsys any, args ...string) {
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, stop := signal.NotifyContext(context.Background(),
		termination.Signals...)
	defer stop()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if len(Input) > 0 {
		f, err := os.Open(Input)
		if err != nil {
			log.Print(err)
			return
		}
		defer f.Close()
		ctx = rctx.Parameter.With(ctx, f)
	}
	if len(Output) > 0 {
		fflags := os.O_RDWR | os.O_CREATE
		if Append {
			fflags |= os.O_APPEND
		} else {
			fflags |= os.O_TRUNC
		}
		fmode := os.FileMode(Mode)
		f, err := os.OpenFile(Output, fflags, fmode)
		if err != nil {
			log.Print(err)
			return
		}
		defer f.Close()
		if Tee {
			mw := io.MultiWriter(wctx.Parameter.In(ctx), f)
			ctx = wctx.Parameter.With(ctx, mw)
		} else {
			ctx = wctx.Parameter.With(ctx, f)
		}
	}
	if Timeout != 0 {
		ctx, cancel = context.WithTimeout(ctx, Timeout)
		defer cancel()
	}
	if show, ok := selctx.Parameter.Default["show"].(map[string]any); ok {
		if _, ok = show["build"]; !ok {
			show["build"] = program.Build
		}
		if _, ok = show["main"]; !ok {
			show["main"] = program.Main
		}
		if _, ok = show["completion"]; !ok {
			show["completion"] = IntegralShowCompletion
		}
	}
	if len(args) > 0 && args[0] == "daemon" && !program.IsKoApp() {
		if wc, err := oslog.OpenError(); err == nil {
			defer wc.Close()
			log.SetOutput(wc)
		}
		if wc, err := oslog.OpenNotice(); err == nil {
			defer wc.Close()
			wctx.Parameter.Default = wc
		}
	}

	err := do(ctx, subsys, args...)
	if err != nil {
		if !IsGoesError(err) &&
			!egress.IsMarked(err) &&
			!usage.InError(err) {

			err = GoesError{pathctx.Parameter.In(ctx), err}
		}
		Fprint1ln(log.Writer(), err)
		if !usage.InError(err) {
			os.Exit(1)
		}
	}
}

func Main() {
	log.SetFlags(log.Lshortfile)
	flag.BoolVar(&Append, "append", false, "Append to -output file.")
	flag.BoolVar(&Append, "a", false, "aka. -append.")
	flag.BoolVar(&Complete, "complete", false, "Finish last arg.")
	flag.BoolVar(&Help, "help", false, "Print options.")
	flag.BoolVar(&Help, "h", false, "aka. -help.")
	flag.BoolVar(&Json, "json", false,
		"Marshal/Unmarshal object with JSON format text.")
	flag.BoolVar(&Tee, "tee", false, "Tee to -output file and stdout.")
	flag.BoolVar(&Tee, "t", false, "aka. -tee.")
	flag.StringVar(&Input, "input", "",
		"Read from named file instead of stdin.")
	flag.StringVar(&Input, "i", "", "aka -i.")
	flag.StringVar(&Output, "output", "",
		"Write to named file instead of stdout.")
	flag.StringVar(&Output, "o", "", "aka. -o")
	flag.UintVar(&Mode, "mode", 0666, "File mode (default 0666).")
	flag.DurationVar(&Timeout, "timeout", 0, "Elapse time limit.")
	flag.Parse()
	selctx.Parameter.Default["integral"] = Integral
	Merge(selctx.Parameter.Default, Integral)
	Exec(Select, flag.Args()...)
}

func Merge(to, from map[string]any) {
	for k, v := range from {
		if _, ok := to[k]; !ok {
			to[k] = v
		}
	}
}

func Select(ctx context.Context, args ...string) (err error) {
	sel := selctx.Parameter.In(ctx)
	if len(sel) == 0 {
		err = ErrEmpty
	} else if len(args) == 0 {
		if Complete {
			err = complete.Last(args, sel)
		} else if Help {
			err = IntegralHelp(ctx)
		} else {
			err = ErrIncomplete
		}
	} else if v, ok := sel[args[0]]; ok {
		ctx = pathctx.AppendIn(ctx, args[0])
		err = do(ctx, v, args[1:]...)
	} else if len(args) == 1 && Complete {
		err = complete.Last(args, sel)
	} else if len(pathctx.Parameter.In(ctx)) == 1 {
		err = IntegralCommand(ctx, args...)
	} else {
		ctx = pathctx.AppendIn(ctx, args[0])
		err = ErrNotFound
	}
	if err != nil &&
		!IsGoesError(err) &&
		!egress.IsMarked(err) &&
		!usage.InError(err) {

		err = GoesError{pathctx.Parameter.In(ctx), err}
	}
	return
}

const MarshalTextUsageTemplate = `
usage: {{.}} [-json]
Format named object.`

func MarshalTextUsageData(ctx context.Context) any {
	return pathctx.StringIn(ctx)
}

const UnmarshalTextUsageTemplate = `
usage : {{.}} [-json] <value>
Unmarshal or scan object from text value.`

var UnmarshalTextUsageData = MarshalTextUsageData

func do(ctx context.Context, subsys any, args ...string) error {
	var text []byte
	var err error
	w := wctx.Parameter.In(ctx)
	switch t := subsys.(type) {
	case []byte:
		if !Json {
			Fprint1ln(w, t)
		} else if text, err = json.Marshal(t); err == nil {
			Fprint1ln(w, text)
		}
		return err
	case string:
		if !Json {
			Fprint1ln(w, t)
		} else if text, err = json.Marshal(t); err == nil {
			Fprint1ln(w, text)
		}
		return err
	case map[string]any:
		return Select(selctx.Parameter.With(ctx, t), args...)
	case embed.FS:
		return IntegralShowFS(ctx, t, args...)
	case func(context.Context, ...string) error:
		return t(ctx, args...)
	case func() ([]byte, error):
		if text, err = t(); err == nil && len(text) > 0 {
			if Json {
				text, err = json.Marshal(text)
			}
			if err == nil && len(text) > 0 {
				Fprint1ln(w, text)
			}
		}
		return err
	case func() string:
		if !Json {
			Fprint1ln(w, t())
		} else if text, err = json.Marshal(t()); err == nil {
			if len(text) > 0 {
				Fprint1ln(w, text)
			}
		}
		return err
	case func() fmt.Stringer:
		s := t().String()
		if !Json {
			Fprint1ln(w, s)
		} else if text, err = json.Marshal(s); err == nil {
			if len(text) > 0 {
				Fprint1ln(w, text)
			}
		}
		return err
	}
	// Show or set objects
	nargs := len(args)
	if Help {
		if nargs == 0 {
			err = usage.Error(MarshalTextUsageTemplate[1:],
				MarshalTextUsageData(ctx))
		} else {
			err = usage.Error(UnmarshalTextUsageTemplate[1:],
				UnmarshalTextUsageData(ctx))
		}
	} else if nargs == 0 {
		if Json {
			text, err = json.MarshalIndent(subsys, "", "  ")
			if err == nil && len(text) > 0 {
				Fprint1ln(w, text)
			}
		} else {
			if tm, ok := subsys.(encoding.TextMarshaler); ok {
				text, err = tm.MarshalText()
				if err == nil && len(text) > 0 {
					Fprint1ln(w, text)
				}
			} else {
				Fprint1ln(w, subsys)
			}
		}
	} else if Json {
		text = []byte(strings.Join(args[1:], " "))
		err = json.Unmarshal(text, subsys)
	} else {
		text = []byte(strings.Join(args, " "))
		if method, ok := subsys.(encoding.TextUnmarshaler); ok {
			err = method.UnmarshalText(text)
		} else {
			_, err = fmt.Sscan(args[0], subsys)
		}
	}
	return err
}

var nl = []byte{'\n'}

// Print args to writer in context with one and only one trailing newline.
// If args[0] is a []byte, write as is instead of fmt.
func Fprint1ln(w io.Writer, args ...any) {
	var b *bytes.Buffer
	if len(args) == 1 {
		switch t := args[0].(type) {
		case []byte:
			b = bytes.NewBuffer(t)
		case string:
			b = bytes.NewBufferString(t)
		default:
			b = new(bytes.Buffer)
			fmt.Fprint(b, args[0])
		}
	} else {
		b = new(bytes.Buffer)
		fmt.Fprint(b, args...)
	}
	hasnl := bytes.HasSuffix(b.Bytes(), nl)
	b.WriteTo(w)
	if !hasnl {
		w.Write(nl)
	}
}
