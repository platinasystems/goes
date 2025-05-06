// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
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
	var gname, euname, egname, s, sep string
	var id_A, id_G, id_M, id_P, id_c, id_g, id_n,
		id_p, id_r, id_u bool

	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [user]
Print “user” (or current user's) identity.

{{flags .}}`)

	xflag.Define(&id_A, "A", "Print user process audit.")
	xflag.Define(&id_G, "G", "Print group IDs.")
	xflag.Define(&id_M, "M", "Print process MAC label.")
	xflag.Define(&id_P, "P", "Print password file entry.")
	xflag.Define(&id_c, "c", "Print login class.")
	xflag.Define(&id_g, "g", "Print effective group ID.")
	xflag.Define(&id_n, "n", "Print user or group name instead of number.")
	xflag.Define(&id_p, "p", "Print human readable output.")
	xflag.Define(&id_r, "r", "Print real user ID.")
	xflag.Define(&id_u, "u", "Print effective user ID.")

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
	case id_A:
		return xerrors.FIXME()
	case id_c:
		return xerrors.FIXME()
	case id_g:
		if len(args) == 0 && !id_r {
			if id_n {
				s = egname
			} else {
				s = segid
			}
		} else if id_n {
			s = gname
		} else {
			s = u.Gid
		}
		fmt.Println(s)
	case id_u:
		if len(args) == 0 && !id_r {
			if id_n {
				s = euname
			} else {
				s = seuid
			}
		} else if id_n {
			s = u.Username
		} else {
			s = u.Uid
		}
		fmt.Println(s)
	case id_G:
		for _, gid := range gids {
			fmt.Print(sep, gid)
			sep = " "
		}
		fmt.Println()
	case id_M:
		return xerrors.FIXME()
	case id_P:
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
	case id_p:
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
