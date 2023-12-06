// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package coreutils

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/ctxparm"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const IdUsageTemplate = `
usage: {{.Path}} [<options>] [<user>],
Print user identity.
{{.Flag}}`

func IdUsageData(ctx context.Context) any {
	return struct{ Path, Flag string }{
		Path: strings.Join(ctxparm.Strings.In(ctx), " "),
		Flag: ctxparm.SprintFlagsIn(ctx),
	}
}

func Id(ctx context.Context, args ...string) error {
	if *complete.Help {
		return nil
	}
	flags := usage.NewFlags("id")
	ctx = ctxparm.Flags.With(ctx, flags)
	Aflag := flags.Bool("A", false, "Print user process audit.")
	Gflag := flags.Bool("G", false, "Print group IDs.")
	Mflag := flags.Bool("M", false, "Print process MAC label.")
	Pflag := flags.Bool("P", false, "Print password file entry.")
	cflag := flags.Bool("c", false, "Print login class.")
	gflag := flags.Bool("g", false, "Print effective group ID.")
	pflag := flags.Bool("p", false, "Print human readable output.")
	uflag := flags.Bool("u", false, "Print effective user ID.")
	nflag := flags.Bool("n", false, "Print user or group name instead of number.")
	rflag := flags.Bool("r", false,
		"Print real instead of effective group or user ID.")
	err := flags.Parse(args)
	if err != nil {
		return err
	}
	if *usage.Help {
		return usage.Error(IdUsageTemplate[1:], IdUsageData(ctx))
	}
	args = flags.Args()
	w := ctxparm.Writer.In(ctx)
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
