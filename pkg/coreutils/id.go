// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package coreutils

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/user"

	"github.com/platinasystems/goes/v2/pkg/goes"
)

const IdUsage = `
usage: {{branch .}} [<options>] [<user>],
Print user identity.
{{flags .}}`

func Id(ctx context.Context, args []string) error {
	var flags flag.FlagSet
	ctx = goes.FlagsContext(ctx, &flags)
	if goes.ContextComplete(ctx) {
		return nil
	}
	Aflag := flags.Bool("A", false, "Print user process audit.")
	Gflag := flags.Bool("G", false, "Print group IDs.")
	Mflag := flags.Bool("M", false, "Print process MAC label.")
	Pflag := flags.Bool("P", false, "Print password file entry.")
	cflag := flags.Bool("c", false, "Print login class.")
	gflag := flags.Bool("g", false, "Print effective group ID.")
	pflag := flags.Bool("p", false, "Print human readable output.")
	uflag := flags.Bool("u", false, "Print effective user ID.")
	nflag := flags.Bool("n", false,
		"Print user or group name instead of number.")
	rflag := flags.Bool("r", false,
		"Print real instead of effective group or user ID.")
	ctx, err := goes.ParseFlagsContext(ctx, args)
	if err != nil {
		return err
	}
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, IdUsage)
	}
	args = flags.Args()
	w := goes.ContextStdout(ctx)
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
