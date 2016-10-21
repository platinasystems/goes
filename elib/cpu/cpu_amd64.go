// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cpu

// Cache lines on x86 are 64 bytes.
const Log2CacheLineBytes = 6

func TimeNow() Time
