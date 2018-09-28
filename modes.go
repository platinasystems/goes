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

import (
	"bytes"
	"fmt"
)

const (
	ETHTOOL_LINK_MODE_10baseT_Half = iota
	ETHTOOL_LINK_MODE_10baseT_Full
	ETHTOOL_LINK_MODE_100baseT_Half
	ETHTOOL_LINK_MODE_100baseT_Full
	ETHTOOL_LINK_MODE_1000baseT_Half
	ETHTOOL_LINK_MODE_1000baseT_Full
	ETHTOOL_LINK_MODE_Autoneg
	ETHTOOL_LINK_MODE_TP
	ETHTOOL_LINK_MODE_AUI
	ETHTOOL_LINK_MODE_MII
	ETHTOOL_LINK_MODE_FIBRE
	ETHTOOL_LINK_MODE_BNC
	ETHTOOL_LINK_MODE_10000baseT_Full
	ETHTOOL_LINK_MODE_Pause
	ETHTOOL_LINK_MODE_Asym_Pause
	ETHTOOL_LINK_MODE_2500baseX_Full
	ETHTOOL_LINK_MODE_Backplane
	ETHTOOL_LINK_MODE_1000baseKX_Full
	ETHTOOL_LINK_MODE_10000baseKX4_Full
	ETHTOOL_LINK_MODE_10000baseKR_Full
	ETHTOOL_LINK_MODE_10000baseR_FEC
	ETHTOOL_LINK_MODE_20000baseMLD2_Full
	ETHTOOL_LINK_MODE_20000baseKR2_Full
	ETHTOOL_LINK_MODE_40000baseKR4_Full
	ETHTOOL_LINK_MODE_40000baseCR4_Full
	ETHTOOL_LINK_MODE_40000baseSR4_Full
	ETHTOOL_LINK_MODE_40000baseLR4_Full
	ETHTOOL_LINK_MODE_56000baseKR4_Full
	ETHTOOL_LINK_MODE_56000baseCR4_Full
	ETHTOOL_LINK_MODE_56000baseSR4_Full
	ETHTOOL_LINK_MODE_56000baseLR4_Full
	ETHTOOL_LINK_MODE_25000baseCR_Full
	ETHTOOL_LINK_MODE_25000baseKR_Full
	ETHTOOL_LINK_MODE_25000baseSR_Full
	ETHTOOL_LINK_MODE_50000baseCR2_Full
	ETHTOOL_LINK_MODE_50000baseKR2_Full
	ETHTOOL_LINK_MODE_100000baseKR4_Full
	ETHTOOL_LINK_MODE_100000baseSR4_Full
	ETHTOOL_LINK_MODE_100000baseCR4_Full
	ETHTOOL_LINK_MODE_100000baseLR4_ER4_Full
	ETHTOOL_LINK_MODE_50000baseSR2_Full
	ETHTOOL_LINK_MODE_1000baseX_Full
	ETHTOOL_LINK_MODE_10000baseCR_Full
	ETHTOOL_LINK_MODE_10000baseSR_Full
	ETHTOOL_LINK_MODE_10000baseLR_Full
	ETHTOOL_LINK_MODE_10000baseLRM_Full
	ETHTOOL_LINK_MODE_10000baseER_Full
	ETHTOOL_LINK_MODE_2500baseT_Full
	ETHTOOL_LINK_MODE_5000baseT_Full
	ETHTOOL_LINK_MODE_NBITS
)

const ETHTOOL_LINK_MODE_NWORDS = (((ETHTOOL_LINK_MODE_NBITS) + 32) - 1) / 32

type EthtoolLinkModeBits [ETHTOOL_LINK_MODE_NWORDS]uint32

func (bits *EthtoolLinkModeBits) Load(from *EthtoolLinkModeBits) {
	copy(bits[:], from[:])
}

func (bits *EthtoolLinkModeBits) Set(n uint) {
	bits[n/32] |= (1 << n)
}

func (bits *EthtoolLinkModeBits) String() string {
	buf := new(bytes.Buffer)
	none := true
	for i, s := range []string{
		"10baseT/Half",
		"10baseT/Full",
		"100baseT/Half",
		"100baseT/Full",
		"1000baseT/Half",
		"1000baseT/Full",
		"Autoneg",
		"TP",
		"AUI",
		"MII",
		"FIBRE",
		"BNC",
		"10000baseT/Full",
		"Pause",
		"Asym-Pause",
		"2500baseX/Full",
		"Backplane",
		"1000baseKX/Full",
		"10000baseKX4/Full",
		"10000baseKR/Full",
		"10000baseR_FEC",
		"20000baseMLD2/Full",
		"20000baseKR2/Full",
		"40000baseKR4/Full",
		"40000baseCR4/Full",
		"40000baseSR4/Full",
		"40000baseLR4/Full",
		"56000baseKR4/Full",
		"56000baseCR4/Full",
		"56000baseSR4/Full",
		"56000baseLR4/Full",
		"25000baseCR/Full",
		"25000baseKR/Full",
		"25000baseSR/Full",
		"50000baseCR2/Full",
		"50000baseKR2/Full",
		"100000baseKR4/Full",
		"100000baseSR4/Full",
		"100000baseCR4/Full",
		"100000baseLR4 ER4/Full",
		"50000baseSR2/Full",
		"1000baseX/Full",
		"10000baseCR/Full",
		"10000baseSR/Full",
		"10000baseLR/Full",
		"10000baseLRM/Full",
		"10000baseER/Full",
		"2500baseT/Full",
		"5000baseT/Full",
	} {
		if bits.Test(uint(i)) {
			fmt.Fprint(buf, "\n\t\t", s)
			none = false
		}
	}
	if none {
		return "\tnone"
	}
	return buf.String()
}

func (bits *EthtoolLinkModeBits) Test(n uint) bool {
	return (bits[n/32] & (1 << n)) != 0
}
