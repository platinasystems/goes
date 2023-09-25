// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

func Netstat(
	ctx context.Context,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `{{$path := join .Path " "}}{{/*
*/}}usage: {{$path}} [<option>]...
Show network status.

  • {{$path}} [-AaLlnW] [-f <family> | -p <protocol>]
  • {{$path}} [-gilns] [-v] [-f <family>] [-I <interface>]
  • {{$path}} -i | -I <interface> [-w <period>] [-c <queue>] [-abdgqRtS]
  • {{$path}} -s [-s] [-f <family> | -p <protocol>] [-w <period>]
  • {{$path}} -i | -I <interface> -s [-f <family> | -p <protocol>]
  • {{$path}} -m [-m]
  • {{$path}} -r [-Aaln] [-f <family>]
  • {{$path}} -rs [-s]
  • {{$path}} -B [-I interface]

Options{{print $.Flags}}`
	fs, h := flag.New()
	iFlag := fs.Bool("i", false, "Show interface info.")
	sFlag := fs.Bool("s", false, "Show per-protocol stats.")
	ssFlag := fs.Bool("ss", false, "Show per-protocol, non-zero stats.")
	mFlag := fs.Bool("m", false, "Show memory stats.")
	mmFlag := fs.Bool("mm", false, "Show detailed memory stats.")
	rFlag := fs.Bool("r", false, "Show routing table.")
	fFlag := fs.String("f", "", "Address Family {inet, inet6, link}.")
	IFlag := fs.String("I", "", "Interface name.")
	pFlag := fs.Int("p", 0, "Protocol number.")
	wFlag := fs.Duration("w", 0, "Wait interval.")
	_ = *sFlag || *ssFlag || *mFlag || *mmFlag || *rFlag
	_ = fFlag
	_ = IFlag
	_ = *pFlag
	_ = *wFlag
	if complete.Parameter.Value(ctx) {
		return nil
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if help.Parameter.Value(ctx) || *h {
		return style.Usage(usage, struct {
			Path  []string
			Flags fmt.Formatter
		}{path, fs})
	}
	args = fs.Args()
	switch {
	case *iFlag:
		nifs, err := netif.List()
		if err != nil {
			return err
		}
		if name := *IFlag; len(name) > 0 {
			for _, nif := range nifs {
				if nif.Name == name {
					nifs[0] = nif
					nifs = nifs[:1]
					break
				}
			}
			if len(nifs) > 1 {
				return fmt.Errorf("%q not found", name)
			}
		}
		fmt.Fprintf(w, "%-15s", "Name")
		fmt.Fprintf(w, " %5s", "MTU")
		fmt.Fprintf(w, " %11s", "Ipkts")
		fmt.Fprintf(w, " %11s", "Ibytes")
		fmt.Fprintf(w, " %11s", "Idrops")
		fmt.Fprintf(w, " %11s", "Ierrs")
		fmt.Fprintf(w, " %11s", "Opkts")
		fmt.Fprintf(w, " %11s", "Obytes")
		fmt.Fprintf(w, " %11s", "Odrops")
		fmt.Fprintf(w, " %11s", "Oerrs")
		fmt.Fprintf(w, " %11s", "Coll")
		fmt.Fprintln(w)
		for _, nif := range nifs {
			fmt.Fprintf(w, "%-15s", nif.Name)
			fmt.Fprintf(w, " %5d", nif.MTU)
			fmt.Fprintf(w, " %11d", nif.Rx.Packets)
			fmt.Fprintf(w, " %11d", nif.Rx.Bytes)
			fmt.Fprintf(w, " %11d", nif.Rx.Drops)
			fmt.Fprintf(w, " %11d", nif.Rx.Errors)
			fmt.Fprintf(w, " %11d", nif.Tx.Packets)
			fmt.Fprintf(w, " %11d", nif.Tx.Bytes)
			fmt.Fprintf(w, " %11d", nif.Tx.Drops)
			fmt.Fprintf(w, " %11d", nif.Tx.Errors)
			fmt.Fprintf(w, " %11d", nif.Collisions)
			fmt.Fprintln(w)
		}
	default:
		return errors.New("FIXME")
	}
	return nil
}
