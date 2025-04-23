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

const (
	Id_A_Flag xflag.KeyUsage[bool] = "A Print user process audit."
	Id_G_Flag xflag.KeyUsage[bool] = "G Print group IDs."
	Id_M_Flag xflag.KeyUsage[bool] = "M Print process MAC label."
	Id_P_Flag xflag.KeyUsage[bool] = "P Print password file entry."
	Id_c_Flag xflag.KeyUsage[bool] = "c Print login class."
	Id_g_Flag xflag.KeyUsage[bool] = "g Print effective group ID."
	Id_p_Flag xflag.KeyUsage[bool] = "p Print human readable output."
	Id_u_Flag xflag.KeyUsage[bool] = "u Print effective user ID."
	Id_n_Flag xflag.KeyUsage[bool] = "n " +
		"Print user or group name instead of number."
	Id_r_Flag xflag.KeyUsage[bool] = "r " +
		"Print real instead of effective group or user ID."
)

func Id(ctx context.Context, args []string) error {
	var u *user.User
	var gname string
	var euname, egname string
	var s, sep string

	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [user]
Print “user” (or current user's) identity.

{{flags .}}`)

	Aflag := Id_A_Flag.Define(false)
	Gflag := Id_G_Flag.Define(false)
	Mflag := Id_M_Flag.Define(false)
	Pflag := Id_P_Flag.Define(false)
	cflag := Id_c_Flag.Define(false)
	gflag := Id_g_Flag.Define(false)
	pflag := Id_p_Flag.Define(false)
	uflag := Id_u_Flag.Define(false)
	nflag := Id_n_Flag.Define(false)
	rflag := Id_r_Flag.Define(false)

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
