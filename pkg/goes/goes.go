// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"bytes"
	"context"
	"embed"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/log/oslog"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/termination"
	"github.com/platinasystems/goes/v2/pkg/override"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const nl = "\n"

var Flag = struct {
	Append, Tee   *bool
	Input, Output *string
	Mode          *uint
	Timeout       *time.Duration
}{
	Append:  flag.Bool("a", false, "Append to -o <file>."),
	Tee:     flag.Bool("t", false, "Tee to -o <file> and stdout."),
	Input:   flag.String("i", "", "Read from named file."),
	Output:  flag.String("o", "", "Write to named file."),
	Mode:    flag.Uint("m", 0, "Output file mode (default 0666)."),
	Timeout: flag.Duration("T", 0, "Elapse time limit."),
}

// Execute subsystem with an interruptible context.
func Exec(ctx context.Context, subsys any, args []string) {
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, stop := signal.NotifyContext(ctx, termination.Signals...)
	defer stop()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := do(ctx, subsys, args)
	if err == nil {
		return
	}
	if !IsWrappedBranch(err) && !IsUsage(err) &&
		!egress.IsMarked(err) {
		err = WrapBranch(ctx, err)
	}
	data := []byte(err.Error())
	w := log.Writer()
	w.Write(bytes.TrimRight(data, nl))
	w.Write([]byte(nl))
	if !IsUsage(err) {
		os.Exit(1)
	}
}

// Parse CommandLine then Exec Select of subsytem.
func Main() {
	log.SetFlags(log.Lshortfile)

	ctx, err := ParseFlagsContext(context.Background(), os.Args[1:])
	if err != nil {
		fmt.Fprint(os.Stderr, ContextBranch(ctx)[0], ": ",
			strings.TrimRight(err.Error(), nl), nl)
		os.Exit(1)
	}
	args := ContextFlags(ctx).Args()

	if Flag.Timeout.Nanoseconds() > 0 {
		var timeout context.CancelFunc
		ctx, timeout = context.WithTimeout(ctx, *Flag.Timeout)
		defer timeout()
	}
	if len(*Flag.Input) > 0 {
		f, err := os.Open(*Flag.Input)
		if err != nil {
			log.Print(err)
			return
		}
		defer f.Close()
		defer override.Value(&os.Stdin, f)()
	}
	if len(*Flag.Output) > 0 {
		fflags := os.O_RDWR | os.O_CREATE
		if *Flag.Append {
			fflags |= os.O_APPEND
		} else {
			fflags |= os.O_TRUNC
		}
		mode := os.FileMode(0666)
		if *Flag.Mode != 0 {
			mode = os.FileMode(*Flag.Mode)
		}
		f, err := os.OpenFile(*Flag.Output, fflags, mode)
		if err != nil {
			log.Print(err)
			return
		}
		defer f.Close()
		if *Flag.Tee {
			pr, pw, err := os.Pipe()
			if err != nil {
				log.Print(err)
				return
			}
			defer pw.Close()
			defer override.Value(&os.Stdout, pw)()
			go io.Copy(io.MultiWriter(os.Stdout, f), pr)
		} else {
			defer override.Value(&os.Stdout, f)()
		}
	}
	if len(args) > 0 && args[0] == "daemon" && !program.IsKoApp() {
		errLog, err := oslog.OpenError()
		if err != nil {
			log.Fatal(err)
		}
		defer errLog.Close()
		log.SetOutput(errLog)
		rErrPipe, wErrPipe, err := os.Pipe()
		if err != nil {
			log.Fatal(err)
		}
		defer wErrPipe.Close()
		defer override.Value(&os.Stderr, wErrPipe)()
		go io.Copy(errLog, rErrPipe)
		if len(*Flag.Output) == 0 {
			outLog, err := oslog.OpenNotice()
			if err != nil {
				log.Fatal(err)
			}
			defer outLog.Close()
			rOutPipe, wOutPipe, err := os.Pipe()
			if err == nil {
				log.Fatal(err)
			}
			defer wOutPipe.Close()
			defer override.Value(&os.Stdout, wOutPipe)()
			go io.Copy(outLog, rOutPipe)
		}
	}
	Exec(ctx, Select, args)
}

// Select subsystem from args.
func Select(ctx context.Context, args []string) (err error) {
	root := ContextRoot(ctx)
	if len(root) == 0 {
		err = ErrEmpty
	} else if len(args) == 0 {
		if ContextComplete(ctx) {
			err = complete.Last(args, root)
		} else {
			err = Help(ctx, args)
		}
	} else if v, ok := root[args[0]]; ok {
		ctx = AppendBranchContext(ctx, args[0])
		err = do(ctx, v, args[1:])
	} else if len(args) == 1 && ContextComplete(ctx) {
		err = complete.Last(args, root)
	} else if len(ContextBranch(ctx)) == 1 {
		err = ExternalCommand(ctx, args)
	} else {
		ctx = AppendBranchContext(ctx, args[0])
		err = ErrNotFound
	}
	if err != nil && !IsWrappedBranch(err) && !IsUsage(err) &&
		!egress.IsMarked(err) {
		err = WrapBranch(ctx, err)
	}
	return
}

func do(ctx context.Context, subsys any, args []string) (err error) {
	switch t := subsys.(type) {
	case map[string]any:
		err = Select(RootContext(ctx, t), args)
	case func(context.Context, []string) error:
		err = t(ctx, args)
	case embed.FS:
		err = PrintEmbedFS(ctx, t, args)
	case []byte:
		err = PrintOrReadBytes(ctx, t, args)
	case string:
		err = PrintString(ctx, t, args)
	case fmt.Stringer:
		err = PrintStringer(ctx, t, args)
	case func() ([]byte, error):
		err = PrintBytesResult(ctx, t, args)
	case func() string:
		err = PrintStringResult(ctx, t, args)
	case jsoner:
		err = PrintOrUnmarshalJSON(ctx, t, args)
	case texter:
		err = PrintOrUnmarshalText(ctx, t, args)
	default:
		err = PrintOrScanObject(ctx, subsys, args)
	}
	return
}
