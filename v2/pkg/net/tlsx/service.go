// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/platinasystems/goes/v2/pkg/context/nbr"
	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/os/page"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/utmpx"
)

var (
	ErrEmptyRequest    = errors.New("empty request")
	ErrExit            = errors.New("exit")
	ServiceReadTimeout = 3 * time.Second
)

// Service connection by parsing TLV request arguements and input upto the zero
// length Break.  This calls f() with a reader that LV decodes any input; a
// writer that LV encodes date to connection; and the decoded arguments.  If
// f() succeeds, this sends the zero length Break to the connection; otherwise,
// this sends an encoded Nack.
func service(
	ctx context.Context,
	conn *tls.Conn,
	path selection.Path,
	f selection.Func,
) error {
	var wg sync.WaitGroup
	defer wg.Wait()
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dec := lv.NewDecoder(poll.With(cctx, conn))
	enc := lv.NewEncoder(write.With(cctx, conn))
	r := io.LimitReader(nil, 0)
	w := io.Writer(enc)
	pg := page.New()
	defer page.Free(pg)

	var (
		args  []string
		err   error
		n     int
		ispty bool
	)
	for i := 0; ; {
		if n, err = dec.Read(pg[i:]); err != nil {
			return fmt.Errorf("decode: %w", err)
		} else if n == 0 {
			if len(args) == 0 {
				return ErrEmptyRequest
			}
			break
		} else if s := string(pg[i : i+n]); s != InputTag {
			args = append(args, s)
			i += n
		} else if len(args) == 0 {
			return ErrEmptyRequest
		} else if args[0] == "pty" {
			ispty = true
			break
		} else {
			ir, flush := newinput(dec)
			r = ir
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-cctx.Done()
				flush()
			}()
			break
		}
	}
	ra := RemoteAddr(conn)
	style.Plain.Notice.Println(ra, args)
	if ispty {
		err = ptycmd(cctx, wg, ra, dec, enc, args[1:])
	} else {
		err = f(cctx, r, w, path, args...)
	}
	switch {
	case errors.Is(err, context.Canceled):
	case errors.Is(err, net.ErrClosed):
	case errors.Is(err, io.EOF):
	case err == nil:
		if ispty {
			err = ErrExit
		}
		enc.Encode(err)
	default:
		err = fmt.Errorf("service %v: %w", path, err)
		enc.Encode(err)
	}
	return err
}

func ptycmd(
	ctx context.Context,
	wg sync.WaitGroup,
	ra string,
	dec lv.Decoding,
	enc lv.Encoding,
	args []string,
) error {
	if n := len(args); n < 5 {
		return fmt.Errorf("missing %s", []string{
			"<rows>",
			"<cols>",
			"<x-pixels>",
			"<y-pixels>",
			"<req>",
		}[n])
	}

	var ws pty.Winsize
	for i, pu := range []*uint16{
		&ws.Rows,
		&ws.Cols,
		&ws.X,
		&ws.Y,
	} {
		if _, err := fmt.Sscan(args[i], pu); err != nil {
			return fmt.Errorf("%q: %w",
				args[i], err)
		}
	}
	args = args[4:]

	ptmx, tty, err := pty.Open()
	if err != nil {
		return fmt.Errorf("pty: %w", err)
	}
	if err = pty.Setsize(tty, &ws); err != nil {
		ptmx.Close()
		return fmt.Errorf("resize: %w", err)
	}

	wg.Add(1)
	go func() {
		ir, flush := newinput(dec)
		io.Copy(ptmx, ir)
		ptmx.Close()
		flush()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		io.Copy(enc, nbr.With(ctx, ptmx))
		wg.Done()
	}()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
	}
	if err = cmd.Start(); err == nil {
		if program.IsSuperUser.Value() {
			line := tty.Name()
			pid := cmd.Process.Pid
			utx := utmpx.NewUserProcess("root", line, ra, pid)
			defer utx.Died()
		}
		err = cmd.Wait()
	}
	return err
}
