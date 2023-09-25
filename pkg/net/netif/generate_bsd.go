// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || freebsd || netbsd || openbsd

package netif

//go:generate sh -c "go tool cgo -godefs -- godefs_bsd.go > znetif_${GOOS}_${GOARCH}.go"
//go:generate stringer -output=zifcap_string_${GOOS}_${GOARCH}.go -type=IFCAP -trimprefix=IFCAP_ .
//go:generate sh -c "go doc syscall.IFF_UP | sed -n -f iff.sed > ziff_${GOOS}_${GOARCH}.go"
//go:generate stringer -output=ziff_string_${GOOS}_${GOARCH}.go -type=IFF -trimprefix=IFF_ .
//go:generate sh -c "go doc syscall.IFT_OTHER | sed -n -f ift.sed > zift_${GOOS}_${GOARCH}.go"
//go:generate stringer -output=zift_string_${GOOS}_${GOARCH}.go -type=IFT -trimprefix=IFT_ .
