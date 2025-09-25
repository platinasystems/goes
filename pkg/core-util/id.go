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

var Id_A = xflag.New[bool]("A", "Print user process audit.", nil)
var Id_G = xflag.New[bool]("G", "Print group IDs.", nil)
var Id_M = xflag.New[bool]("M", "Print process MAC label.", nil)
var Id_P = xflag.New[bool]("P", "Print password file entry.", nil)
var Id_c = xflag.New[bool]("c", "Print login class.", nil)
var Id_g = xflag.New[bool]("g", "Print effective group ID.", nil)
var Id_n = xflag.New[bool]("n", "Print user or group name instead of number.",
	nil)
var Id_p = xflag.New[bool]("p", "Print human readable output.", nil)
var Id_r = xflag.New[bool]("r", "Print real user ID.", nil)
var Id_u = xflag.New[bool]("u", "Print effective user ID.", nil)

var IdFlags = []xflag.Definer{
	Id_A,
	Id_G,
	Id_M,
	Id_P,
	Id_c,
	Id_g,
	Id_n,
	Id_p,
	Id_r,
	Id_u,
}

func Id(ctx context.Context, args []string) error {
	var u *user.User
	var gname, euname, egname, s, sep string

	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [user]
Print “user” (or current user's) identity.

{{flags .}}`)

	for _, f := range IdFlags {
		f.Define()
	}

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
	case Id_A.Value():
		return xerrors.FIXME()
	case Id_c.Value():
		return xerrors.FIXME()
	case Id_g.Value():
		if len(args) == 0 && !Id_r.Value() {
			if Id_n.Value() {
				s = egname
			} else {
				s = segid
			}
		} else if Id_n.Value() {
			s = gname
		} else {
			s = u.Gid
		}
		fmt.Println(s)
	case Id_u.Value():
		if len(args) == 0 && !Id_r.Value() {
			if Id_n.Value() {
				s = euname
			} else {
				s = seuid
			}
		} else if Id_n.Value() {
			s = u.Username
		} else {
			s = u.Uid
		}
		fmt.Println(s)
	case Id_G.Value():
		for _, gid := range gids {
			fmt.Print(sep, gid)
			sep = " "
		}
		fmt.Println()
	case Id_M.Value():
		return xerrors.FIXME()
	case Id_P.Value():
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
	case Id_p.Value():
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
