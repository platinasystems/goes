/* Copyright(c) 2018 Platina Systems, Inc.
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms and conditions of the GNU General Public License,
 * version 2, as published by the Free Software Foundation.
 *
 * This program is distributed in the hope it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
 * more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * this program; if not, write to the Free Software Foundation, Inc.,
 * 51 Franklin St - Fifth Floor, Boston, MA 02110-1301 USA.
 *
 * The full GNU General Public License is included in this distribution in
 * the file called "COPYING".
 *
 * Contact Information:
 * sw@platina.com
 * Platina Systems, 3180 Del La Cruz Blvd, Santa Clara, CA 95054
 */
package xeth

const (
	IndexofNetStatRxPackets uint64 = iota
	IndexofNetStatTxPackets
	IndexofNetStatRxBytes
	IndexofNetStatTxBytes
	IndexofNetStatRxErrors
	IndexofNetStatTxErrors
	IndexofNetStatRxDropped
	IndexofNetStatTxDropped
	IndexofNetStatMulticast
	IndexofNetStatCollisions
	IndexofNetStatRxLengthErrors
	IndexofNetStatRxOverErrors
	IndexofNetStatRxCrcErrors
	IndexofNetStatRxFrameErrors
	IndexofNetStatRxFifoErrors
	IndexofNetStatRxMissedErrors
	IndexofNetStatTxAbortedErrors
	IndexofNetStatTxCarrierErrors
	IndexofNetStatTxFifoErrors
	IndexofNetStatTxHeartbeatErrors
	IndexofNetStatTxWindowErrors
	IndexofNetStatRxCompressed
	IndexofNetStatTxCompressed
	IndexofNetStatRxNohandler
)

var IndexofNetStat = map[string]uint64{
	"rx-packets":          IndexofNetStatRxPackets,
	"tx-packets":          IndexofNetStatTxPackets,
	"rx-bytes":            IndexofNetStatRxBytes,
	"tx-bytes":            IndexofNetStatTxBytes,
	"rx-errors":           IndexofNetStatRxErrors,
	"tx-errors":           IndexofNetStatTxErrors,
	"rx-dropped":          IndexofNetStatRxDropped,
	"tx-dropped":          IndexofNetStatTxDropped,
	"multicast":           IndexofNetStatMulticast,
	"collisions":          IndexofNetStatCollisions,
	"rx-length-errors":    IndexofNetStatRxLengthErrors,
	"rx-over-errors":      IndexofNetStatRxOverErrors,
	"rx-crc-errors":       IndexofNetStatRxCrcErrors,
	"rx-frame-errors":     IndexofNetStatRxFrameErrors,
	"rx-fifo-errors":      IndexofNetStatRxFifoErrors,
	"rx-missed-errors":    IndexofNetStatRxMissedErrors,
	"tx-aborted-errors":   IndexofNetStatTxAbortedErrors,
	"tx-carrier-errors":   IndexofNetStatTxCarrierErrors,
	"tx-fifo-errors":      IndexofNetStatTxFifoErrors,
	"tx-heartbeat-errors": IndexofNetStatTxHeartbeatErrors,
	"tx-window-errors":    IndexofNetStatTxWindowErrors,
	"rx-compressed":       IndexofNetStatRxCompressed,
	"tx-compressed":       IndexofNetStatTxCompressed,
	"rx-nohandler":        IndexofNetStatRxNohandler,
}
