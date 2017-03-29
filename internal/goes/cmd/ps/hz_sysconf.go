// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build cgo,!netgo

package ps

/*
   #include <unistd.h>
   #include <sys/types.h>
*/
import "C"

func Hz() uint64 { return uint64(C.sysconf(C._SC_CLK_TCK)) }
