// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin

package netif

import (
	"context"
	"fmt"
	"net/netip"
	"os/exec"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

const ND6_INFINITE_LIFETIME = 0xffffffff

const AddressCommands = `
  add	Add (default) network address to network interface. (aka. alias)
  del	Remove network address to network interface. (aka. -alias)
`

const AddressParameters = ""

func (nif *NetIf) Add(
	ctx context.Context,
	addr, dest netip.Addr,
	bits int,
	parms ...string,
) error {
	return nif.ifconfig(ctx, addr, dest, bits, append(parms, "add"))
}

func (nif *NetIf) Change(
	ctx context.Context,
	addr, dest netip.Addr,
	bits int,
	parms ...string,
) error {
	return ErrUnsupported
}

func (nif *NetIf) Del(
	ctx context.Context,
	addr, dest netip.Addr,
	bits int,
	parms ...string,
) error {
	return nif.ifconfig(ctx, addr, dest, bits, append(parms, "delete"))
}

func (nif *NetIf) Replace(
	ctx context.Context,
	addr, dest netip.Addr,
	bits int,
	parms ...string,
) error {
	return ErrUnsupported
}

func (nif *NetIf) ifconfig(
	ctx context.Context,
	addr, dest netip.Addr,
	bits int,
	parms []string,
) error {
	args := []string{nif.Name}

	if addr.Is6() {
		args = append(args, "inet6")
	} else if addr.Is6() {
		args = append(args, "inet")
	}

	sa := fmt.Sprintf("%s/%d", addr.WithZone(""), bits)
	if zone := addr.Zone(); len(zone) > 0 {
		sa += "%" + zone
	}
	args = append(args, sa)

	if dest.IsValid() {
		args = append(args, dest.String())
	}
	args = append(args, parms...)

	cmd := exec.CommandContext(ctx, "ifconfig", args...)
	err := cmd.Run()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			err = xerrors.Label(err, cmd.Args, ee.Stderr)
		} else {
			err = xerrors.Label(err, cmd.Args)
		}
	}
	return err
}
