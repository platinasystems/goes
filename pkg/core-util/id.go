// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package core_util

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/user"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
)

func Id(ctx context.Context, args []string) error {
	var u *user.User
	var gname string
	var euname, egname string
	var s, sep string

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] [user]
Print “user” (or current user's) identity.

{{flags .}}`)

	Aflag := flag.Bool("A", false, "Print user process audit.")
	Gflag := flag.Bool("G", false, "Print group IDs.")
	Mflag := flag.Bool("M", false, "Print process MAC label.")
	Pflag := flag.Bool("P", false, "Print password file entry.")
	cflag := flag.Bool("c", false, "Print login class.")
	gflag := flag.Bool("g", false, "Print effective group ID.")
	pflag := flag.Bool("p", false, "Print human readable output.")
	uflag := flag.Bool("u", false, "Print effective user ID.")
	nflag := flag.Bool("n", false,
		"Print user or group name instead of number.")
	rflag := flag.Bool("r", false,
		"Print real instead of effective group or user ID.")

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	} else if args = flag.Args(); len(args) == 0 {
		if u, err = user.Current(); err != nil {
			return err
		}
	} else if u, err = user.Lookup(args[0]); err != nil {
		if u, err = user.LookupId(args[0]); err != nil {
			return err
		}
	}

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

	switch {
	case *Aflag:
		return xerrors.FIXME()
	case *cflag:
		return xerrors.FIXME()
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
		fmt.Println(s)
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
		fmt.Println(s)
	case *Gflag:
		for _, gid := range gids {
			fmt.Print(sep, gid)
			sep = " "
		}
		fmt.Println()
	case *Mflag:
		return xerrors.FIXME()
	case *Pflag:
		shell := os.Getenv("SHELL")
		if len(shell) == 0 || len(args) > 0 {
			shell = "SHELL"
		}
		fmt.Print(u.Username,
			":*",
			":", u.Uid,
			":", u.Gid,
			"::0:0",
			":", u.Name,
			":", u.HomeDir,
			":", shell,
			"\n")
	case *pflag:
		fmt.Print("uid\t", u.Username, "\n")
		if len(args) == 0 {
			if seuid != u.Uid {
				fmt.Print("euid\t", seuid, "\n")
			}
			if segid != u.Gid {
				fmt.Print("egid\t", segid, "\n")
			}
		}
		fmt.Printf("groups\t%")
		for _, g := range groups {
			if g != nil {
				fmt.Print(sep, g.Name)
				sep = " "
			}
		}
		fmt.Println()
	default:
		fmt.Printf("uid=%s(%s)", u.Uid, u.Username)
		fmt.Printf(" gid=%s(%s)", u.Gid, gname)
		fmt.Printf(" groups=")
		for _, g := range groups {
			if g != nil {
				fmt.Print(sep)
				sep = ","
				fmt.Printf("%s(%s)", g.Gid, g.Name)
			}
		}
		fmt.Println()
	}
	return ctx.Err()
}
