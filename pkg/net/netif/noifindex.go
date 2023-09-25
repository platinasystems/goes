// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !linux

package netif

const SIOCGIFINDEX = 0

type IFIndex struct{}
