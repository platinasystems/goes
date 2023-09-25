// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

//go:generate sh -c "go doc syscall.IFF_UP | sed -n -f iff.sed > ziff_linux_${GOARCH}.go"
//go:generate stringer -output=ziff_string_linux_${GOARCH}.go -type=IFF -trimprefix=IFF_ .
//go:generate sh -c "go doc syscall.ARPHRD_ETHER | sed -n -f arphrd.sed > zarphrd_linux_${GOARCH}.go"
//go:generate stringer -output=zarphrd_string_linux_${GOARCH}.go -type=ARPHRD -trimprefix=ARPHRD_ .
