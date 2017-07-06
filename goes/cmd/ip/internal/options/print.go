// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import (
	"fmt"
	"io"
	"os"

	"github.com/platinasystems/go/goes/cmd/ip/internal/group"
)

type Newline struct{}
type Gid uint32
type Stat uint64

// Print args then pad with spaces to at least N runes.
// If the printed args were more than N runes, pad with 1 space.
func (opt *Options) Nprint(n int, args ...interface{}) (int, error) {
	const pad = "                                                        "
	i, err := opt.Print(args...)
	if err != nil {
		return i, err
	}
	n -= i
	switch {
	case n <= 0:
		n = 1
	case n > len(pad):
		n = len(pad)
	}
	_, err = fmt.Print(pad[:n])
	return i + n, err
}

func (opt *Options) Print(args ...interface{}) (int, error) {
	var (
		n, total int
		err      error
	)
	for _, v := range args {
		if n, err = opt.Vprint(v); err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

func (opt *Options) Println(args ...interface{}) (int, error) {
	var (
		n, total int
		err      error
	)
	for i, v := range args {
		if n, err = opt.Vprint(v); err != nil {
			return total, err
		}
		total += n
		if i < len(args)-1 {
			if n, err = os.Stdout.Write([]byte{' '}); err != nil {
				return total, err
			}
			total += n
		}
	}
	if opt.Flags.ByName["-o"] {
		n, err = os.Stdout.Write([]byte{'\\'})
	} else {
		n, err = os.Stdout.Write([]byte{'\n'})
	}
	if err == nil {
		total += n
	}
	return total, err
}

func (opt *Options) Vprint(v interface{}) (n int, err error) {
	switch t := v.(type) {
	case []byte:
		n, err = os.Stdout.Write(t)
	case string:
		n, err = os.Stdout.WriteString(t)
	case Gid:
		n, err = os.Stdout.WriteString(group.Name(uint32(t)))
	case Stat:
		if opt.Flags.ByName["-h"] {
			f := float64(t)
			if opt.Flags.ByName["-iec"] {
				const (
					Ki  = 1 << 10
					Mi  = 1 << 20
					Gi  = 1 << 30
					Ti  = 1 << 40
					Pi  = 1 << 50
					Ei  = 1 << 60
					Kif = float64(Ki)
					Mif = float64(Mi)
					Gif = float64(Gi)
					Pif = float64(Ti)
					Eif = float64(Ei)
				)
				switch {
				case t < Ki:
					n, err = fmt.Print(t)
				case t < Mi:
					n, err = fmt.Printf("%.1fKi", f/Kif)
				case t < Gi:
					n, err = fmt.Printf("%.1fMi", f/Mif)
				case t < Ti:
					n, err = fmt.Printf("%.1fGi", f/Gif)
				case t < Pi:
					n, err = fmt.Printf("%.1fTi", f/Pif)
				case t < Ei:
					n, err = fmt.Printf("%.1fPi", f/Pif)
				default:
					n, err = fmt.Printf("%.1fEi", f/Eif)
				}
			} else {
				const (
					K  = 1000
					M  = 1000000
					G  = 1000000000
					T  = 1000000000000
					P  = 1000000000000000
					E  = 1000000000000000000
					Kf = float64(K)
					Mf = float64(M)
					Gf = float64(G)
					Tf = float64(T)
					Pf = float64(P)
					Ef = float64(E)
				)
				switch {
				case t < K:
					n, err = fmt.Print(t)
				case t < M:
					n, err = fmt.Printf("%.1fK", f/Kf)
				case t < G:
					n, err = fmt.Printf("%.1fM", f/Mf)
				case t < T:
					n, err = fmt.Printf("%.1fG", f/Gf)
				case t < P:
					n, err = fmt.Printf("%.1fT", f/Tf)
				case t < E:
					n, err = fmt.Printf("%.1fP", f/Pf)
				default:
					n, err = fmt.Printf("%.1fE", f/Ef)
				}
			}
		} else {
			n, err = fmt.Print(t)
		}
	default:
		if method, found := v.(io.WriterTo); found {
			var i64 int64
			i64, err = method.WriteTo(os.Stdout)
			n = int(i64)
		} else if method, found := v.(fmt.Stringer); found {
			n, err = os.Stdout.WriteString(method.String())
		} else {
			n, err = fmt.Print(v)
		}
	}
	return
}
