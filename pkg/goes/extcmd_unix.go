// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"fmt"
	"os/user"

	"golang.org/x/sys/unix"
)

func NewCredential(groups ...uint32) (*Credential, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}
	credential := new(Credential)
	if _, err = fmt.Sscan(u.Uid, &credential.Uid); err != nil {
		return nil, fmt.Errorf("user:uid: %w", err)
	}
	if _, err = fmt.Sscan(u.Gid, &credential.Gid); err != nil {
		return nil, fmt.Errorf("user:gid: %w", err)
	}
	if len(groups) > 0 {
		credential.Groups = groups
	} else {
		credential.NoSetGroups = true
	}
	return credential, nil
}

func DaemonSysProcAttr() (*unix.SysProcAttr, error) {
	credential, err := NewCredential()
	if err != nil {
		return nil, err
	}
	return &unix.SysProcAttr{
		Credential: credential,
		Setsid:     true,
	}, nil
}
