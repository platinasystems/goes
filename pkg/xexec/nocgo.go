// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !cgo

package xexec

type UserProcess struct{}

func NewUserProcess(user, line, host string, pid int) UserProcess {
	return UserProcess{}
}

func (UserProcess) Died() {}
