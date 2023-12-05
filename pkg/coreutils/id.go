// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package coreutils

import (
	"context"
	"fmt"
	"os"
	"os/user"

	"github.com/platinasystems/goes/v2/pkg/context/flagctx"
	"github.com/platinasystems/goes/v2/pkg/context/pathctx"
	"github.com/platinasystems/goes/v2/pkg/context/wctx"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const IdUsageTemplate = `
usage: {{.Path}} [<options>] [<user>],
Print user identity.
{{.Flag}}`

func IdUsageData(ctx context.Context) any {
	return struct{ Path, Flag string }{
		Path: pathctx.StringIn(ctx),
		Flag: flagctx.StringIn(ctx),
	}
}

func Id(ctx context.Context, args ...string) error {
	fs := flag.NewSilentFlagSet("id")
	ctx = flagctx.Parameter.With(ctx, fs)
	Aflag := fs.Bool("A", false, "Print user process audit.")
	Gflag := fs.Bool("G", false, "Print group IDs.")
	Mflag := fs.Bool("M", false, "Print process MAC label.")
	Pflag := fs.Bool("P", false, "Print password file entry.")
	cflag := fs.Bool("c", false, "Print login class.")
	gflag := fs.Bool("g", false, "Print effective group ID.")
	pflag := fs.Bool("p", false, "Print human readable output.")
	uflag := fs.Bool("u", false, "Print effective user ID.")
	nflag := fs.Bool("n", false, "Print user or group name instead of number.")
	rflag := fs.Bool("r", false,
		"Print real instead of effective group or user ID.")
	if flag.Search[bool]("complete") {
		return complete.Last(args, fs)
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if flag.Search[bool]("help", fs) {
		return usage.Error(IdUsageTemplate[1:], IdUsageData(ctx))
	}
	args = fs.Args()
	w := wctx.Parameter.In(ctx)
	var u *user.User
	if len(args) == 0 {
		if u, err = user.Current(); err != nil {
			return err
		}
	} else if u, err = user.Lookup(args[0]); err != nil {
		if u, err = user.LookupId(args[0]); err != nil {
			return err
		}
	}
	var gname string
	if g, err := user.LookupGroupId(u.Gid); err != nil {
		gname = u.Gid
	} else {
		gname = g.Name
	}
	gids, err := u.GroupIds()
	if err != nil {
		return err
	}
	groups := make([]*user.Group, len(gids))
	for i, gid := range gids {
		groups[i], _ = user.LookupGroupId(gid)
	}
	euid, egid := os.Geteuid(), os.Getegid()
	seuid, segid := fmt.Sprint(euid), fmt.Sprint(egid)
	var euname, egname string
	if eu, err := user.LookupId(seuid); err != nil {
		euname = seuid
	} else {
		euname = eu.Username
	}
	if eg, err := user.LookupGroupId(segid); err != nil {
		egname = segid
	} else {
		egname = eg.Name
	}
	var s, sep string
	switch {
	case *Aflag:
		return FIXME
	case *cflag:
		return FIXME
	case *gflag:
		if len(args) == 0 && !*rflag {
			if *nflag {
				s = egname
			} else {
				s = segid
			}
		} else if *nflag {
			s = gname
		} else {
			s = u.Gid
		}
		fmt.Fprintln(w, s)
	case *uflag:
		if len(args) == 0 && !*rflag {
			if *nflag {
				s = euname
			} else {
				s = seuid
			}
		} else if *nflag {
			s = u.Username
		} else {
			s = u.Uid
		}
		fmt.Fprintln(w, s)
	case *Gflag:
		for _, gid := range gids {
			fmt.Fprint(w, sep, gid)
			sep = " "
		}
		fmt.Fprintln(w)
	case *Mflag:
		return FIXME
	case *Pflag:
		shell := os.Getenv("SHELL")
		if len(shell) == 0 || len(args) > 0 {
			shell = "SHELL"
		}
		fmt.Fprint(w, u.Username,
			":*",
			":", u.Uid,
			":", u.Gid,
			"::0:0",
			":", u.Name,
			":", u.HomeDir,
			":", shell,
			"\n")
	case *pflag:
		fmt.Fprint(w, "uid\t", u.Username, "\n")
		if len(args) == 0 {
			if seuid != u.Uid {
				fmt.Fprint(w, "euid\t", seuid, "\n")
			}
			if segid != u.Gid {
				fmt.Fprint(w, "egid\t", segid, "\n")
			}
		}
		fmt.Fprintf(w, "groups\t%")
		for _, g := range groups {
			if g != nil {
				fmt.Fprint(w, sep, g.Name)
				sep = " "
			}
		}
		fmt.Fprintln(w)
	default:
		fmt.Fprintf(w, "uid=%s(%s)", u.Uid, u.Username)
		fmt.Fprintf(w, " gid=%s(%s)", u.Gid, gname)
		fmt.Fprintf(w, " groups=")
		for _, g := range groups {
			if g != nil {
				fmt.Fprint(w, sep)
				sep = ","
				fmt.Fprintf(w, "%s(%s)", g.Gid, g.Name)
			}
		}
		fmt.Fprintln(w)
	}
	return ctx.Err()
}
