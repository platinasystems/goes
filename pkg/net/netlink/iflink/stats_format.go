// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package iflink

import "fmt"

func (stats Stats[T]) Format(w fmt.State, verb rune) {
	var comma string
	if stats.RxPackets != 0 {
		fmt.Fprint(w, comma, "rx-pkts:", stats.RxPackets)
		comma = ","
	}
	if stats.RxBytes != 0 {
		fmt.Fprint(w, comma, "rx-bytes:", stats.RxBytes)
		comma = ","
	}
	if stats.RxDropped != 0 {
		fmt.Fprint(w, comma, "rx-dropped:", stats.RxDropped)
		comma = ","
	}
	if stats.RxErrors != 0 {
		fmt.Fprint(w, comma, "rx-errors:", stats.RxErrors)
		comma = ","
	}
	if stats.TxPackets != 0 {
		fmt.Fprint(w, comma, "tx-pkts:", stats.TxPackets)
		comma = ","
	}
	if stats.TxBytes != 0 {
		fmt.Fprint(w, comma, "tx-bytes:", stats.TxBytes)
		comma = ","
	}
	if stats.TxDropped != 0 {
		fmt.Fprint(w, comma, "tx-dropped:", stats.TxDropped)
		comma = ","
	}
	if stats.TxErrors != 0 {
		fmt.Fprint(w, comma, "tx-errors:", stats.TxErrors)
		comma = ","
	}
}
