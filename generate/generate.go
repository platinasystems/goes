// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package generate builds goes machines
package generate

//go:generate mkdir -p ../bin
//go:generate echo version...
//go:generate go generate github.com/platinasystems/go/version
//go:generate echo bin/goesd-example...
//go:generate go build -o ../bin/goesd-example github.com/platinasystems/go/goes/goesd-example
//go:generate echo bin/goes-example...
//go:generate go build -o ../bin/goes-example -tags netgo -ldflags -d github.com/platinasystems/go/goes/goes-example
//go:generate echo bin/goesd-platina-mk1...
//go:generate go build -o ../bin/goesd-platina-mk1 github.com/platinasystems/go/goes/goesd-platina-mk1
//go:generate echo bin/goes-platina-mk1-bmc...
//go:generate env GOARCH=arm go build -o ../bin/goes-platina-mk1 -tags netgo -ldflags -d github.com/platinasystems/go/goes/goes-platina-mk1-bmc
