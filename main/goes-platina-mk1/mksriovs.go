// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import . "github.com/platinasystems/go/internal/sriovs"

var (
	eth_0_0  = Vf(Port(0) | SubPort(0) | Vlan(6))
	eth_0_1  = Vf(Port(0) | SubPort(1) | Vlan(7))
	eth_0_2  = Vf(Port(0) | SubPort(2) | Vlan(8))
	eth_0_3  = Vf(Port(0) | SubPort(3) | Vlan(9))
	eth_1_0  = Vf(Port(1) | SubPort(0) | Vlan(2))
	eth_1_1  = Vf(Port(1) | SubPort(1) | Vlan(3))
	eth_1_2  = Vf(Port(1) | SubPort(2) | Vlan(4))
	eth_1_3  = Vf(Port(1) | SubPort(3) | Vlan(5))
	eth_2_0  = Vf(Port(2) | SubPort(0) | Vlan(14))
	eth_2_1  = Vf(Port(2) | SubPort(1) | Vlan(15))
	eth_2_2  = Vf(Port(2) | SubPort(2) | Vlan(16))
	eth_2_3  = Vf(Port(2) | SubPort(3) | Vlan(17))
	eth_3_0  = Vf(Port(3) | SubPort(0) | Vlan(10))
	eth_3_1  = Vf(Port(3) | SubPort(1) | Vlan(11))
	eth_3_2  = Vf(Port(3) | SubPort(2) | Vlan(12))
	eth_3_3  = Vf(Port(3) | SubPort(3) | Vlan(13))
	eth_4_0  = Vf(Port(4) | SubPort(0) | Vlan(22))
	eth_4_1  = Vf(Port(4) | SubPort(1) | Vlan(23))
	eth_4_2  = Vf(Port(4) | SubPort(2) | Vlan(24))
	eth_4_3  = Vf(Port(4) | SubPort(3) | Vlan(25))
	eth_5_0  = Vf(Port(5) | SubPort(0) | Vlan(18))
	eth_5_1  = Vf(Port(5) | SubPort(1) | Vlan(19))
	eth_5_2  = Vf(Port(5) | SubPort(2) | Vlan(20))
	eth_5_3  = Vf(Port(5) | SubPort(3) | Vlan(21))
	eth_6_0  = Vf(Port(6) | SubPort(0) | Vlan(30))
	eth_6_1  = Vf(Port(6) | SubPort(1) | Vlan(31))
	eth_6_2  = Vf(Port(6) | SubPort(2) | Vlan(32))
	eth_6_3  = Vf(Port(6) | SubPort(3) | Vlan(33))
	eth_7_0  = Vf(Port(7) | SubPort(0) | Vlan(26))
	eth_7_1  = Vf(Port(7) | SubPort(1) | Vlan(27))
	eth_7_2  = Vf(Port(7) | SubPort(2) | Vlan(28))
	eth_7_3  = Vf(Port(7) | SubPort(3) | Vlan(29))
	eth_8_0  = Vf(Port(8) | SubPort(0) | Vlan(38))
	eth_8_1  = Vf(Port(8) | SubPort(1) | Vlan(39))
	eth_8_2  = Vf(Port(8) | SubPort(2) | Vlan(40))
	eth_8_3  = Vf(Port(8) | SubPort(3) | Vlan(41))
	eth_9_0  = Vf(Port(9) | SubPort(0) | Vlan(34))
	eth_9_1  = Vf(Port(9) | SubPort(1) | Vlan(35))
	eth_9_2  = Vf(Port(9) | SubPort(2) | Vlan(36))
	eth_9_3  = Vf(Port(9) | SubPort(3) | Vlan(37))
	eth_10_0 = Vf(Port(10) | SubPort(0) | Vlan(46))
	eth_10_1 = Vf(Port(10) | SubPort(1) | Vlan(47))
	eth_10_2 = Vf(Port(10) | SubPort(2) | Vlan(48))
	eth_10_3 = Vf(Port(10) | SubPort(3) | Vlan(49))
	eth_11_0 = Vf(Port(11) | SubPort(0) | Vlan(42))
	eth_11_1 = Vf(Port(11) | SubPort(1) | Vlan(43))
	eth_11_2 = Vf(Port(11) | SubPort(2) | Vlan(44))
	eth_11_3 = Vf(Port(11) | SubPort(3) | Vlan(45))
	eth_12_0 = Vf(Port(12) | SubPort(0) | Vlan(54))
	eth_12_1 = Vf(Port(12) | SubPort(1) | Vlan(55))
	eth_12_2 = Vf(Port(12) | SubPort(2) | Vlan(56))
	eth_12_3 = Vf(Port(12) | SubPort(3) | Vlan(57))
	eth_13_0 = Vf(Port(13) | SubPort(0) | Vlan(50))
	eth_13_1 = Vf(Port(13) | SubPort(1) | Vlan(51))
	eth_13_2 = Vf(Port(13) | SubPort(2) | Vlan(52))
	eth_13_3 = Vf(Port(13) | SubPort(3) | Vlan(53))
	eth_14_0 = Vf(Port(14) | SubPort(0) | Vlan(62))
	eth_14_1 = Vf(Port(14) | SubPort(1) | Vlan(63))
	eth_14_2 = Vf(Port(14) | SubPort(2) | Vlan(64))
	eth_14_3 = Vf(Port(14) | SubPort(3) | Vlan(65))
	eth_15_0 = Vf(Port(15) | SubPort(0) | Vlan(58))
	eth_15_1 = Vf(Port(15) | SubPort(1) | Vlan(59))
	eth_15_2 = Vf(Port(15) | SubPort(2) | Vlan(60))
	eth_15_3 = Vf(Port(15) | SubPort(3) | Vlan(61))
	eth_16_0 = Vf(Port(16) | SubPort(0) | Vlan(70))
	eth_16_1 = Vf(Port(16) | SubPort(1) | Vlan(71))
	eth_16_2 = Vf(Port(16) | SubPort(2) | Vlan(72))
	eth_16_3 = Vf(Port(16) | SubPort(3) | Vlan(73))
	eth_17_0 = Vf(Port(17) | SubPort(0) | Vlan(66))
	eth_17_1 = Vf(Port(17) | SubPort(1) | Vlan(67))
	eth_17_2 = Vf(Port(17) | SubPort(2) | Vlan(68))
	eth_17_3 = Vf(Port(17) | SubPort(3) | Vlan(69))
	eth_18_0 = Vf(Port(18) | SubPort(0) | Vlan(78))
	eth_18_1 = Vf(Port(18) | SubPort(1) | Vlan(79))
	eth_18_2 = Vf(Port(18) | SubPort(2) | Vlan(80))
	eth_18_3 = Vf(Port(18) | SubPort(3) | Vlan(81))
	eth_19_0 = Vf(Port(19) | SubPort(0) | Vlan(74))
	eth_19_1 = Vf(Port(19) | SubPort(1) | Vlan(75))
	eth_19_2 = Vf(Port(19) | SubPort(2) | Vlan(76))
	eth_19_3 = Vf(Port(19) | SubPort(3) | Vlan(77))
	eth_20_0 = Vf(Port(20) | SubPort(0) | Vlan(86))
	eth_20_1 = Vf(Port(20) | SubPort(1) | Vlan(87))
	eth_20_2 = Vf(Port(20) | SubPort(2) | Vlan(88))
	eth_20_3 = Vf(Port(20) | SubPort(3) | Vlan(89))
	eth_21_0 = Vf(Port(21) | SubPort(0) | Vlan(82))
	eth_21_1 = Vf(Port(21) | SubPort(1) | Vlan(83))
	eth_21_2 = Vf(Port(21) | SubPort(2) | Vlan(84))
	eth_21_3 = Vf(Port(21) | SubPort(3) | Vlan(85))
	eth_22_0 = Vf(Port(22) | SubPort(0) | Vlan(94))
	eth_22_1 = Vf(Port(22) | SubPort(1) | Vlan(95))
	eth_22_2 = Vf(Port(22) | SubPort(2) | Vlan(96))
	eth_22_3 = Vf(Port(22) | SubPort(3) | Vlan(97))
	eth_23_0 = Vf(Port(23) | SubPort(0) | Vlan(90))
	eth_23_1 = Vf(Port(23) | SubPort(1) | Vlan(91))
	eth_23_2 = Vf(Port(23) | SubPort(2) | Vlan(92))
	eth_23_3 = Vf(Port(23) | SubPort(3) | Vlan(93))
	eth_24_0 = Vf(Port(24) | SubPort(0) | Vlan(102))
	eth_24_1 = Vf(Port(24) | SubPort(1) | Vlan(103))
	eth_24_2 = Vf(Port(24) | SubPort(2) | Vlan(104))
	eth_24_3 = Vf(Port(24) | SubPort(3) | Vlan(105))
	eth_25_0 = Vf(Port(25) | SubPort(0) | Vlan(98))
	eth_25_1 = Vf(Port(25) | SubPort(1) | Vlan(99))
	eth_25_2 = Vf(Port(25) | SubPort(2) | Vlan(100))
	eth_25_3 = Vf(Port(25) | SubPort(3) | Vlan(101))
	eth_26_0 = Vf(Port(26) | SubPort(0) | Vlan(110))
	eth_26_1 = Vf(Port(26) | SubPort(1) | Vlan(111))
	eth_26_2 = Vf(Port(26) | SubPort(2) | Vlan(112))
	eth_26_3 = Vf(Port(26) | SubPort(3) | Vlan(113))
	eth_27_0 = Vf(Port(27) | SubPort(0) | Vlan(106))
	eth_27_1 = Vf(Port(27) | SubPort(1) | Vlan(107))
	eth_27_2 = Vf(Port(27) | SubPort(2) | Vlan(108))
	eth_27_3 = Vf(Port(27) | SubPort(3) | Vlan(109))
	eth_28_0 = Vf(Port(28) | SubPort(0) | Vlan(118))
	eth_28_1 = Vf(Port(28) | SubPort(1) | Vlan(119))
	eth_28_2 = Vf(Port(28) | SubPort(2) | Vlan(120))
	eth_28_3 = Vf(Port(28) | SubPort(3) | Vlan(121))
	eth_29_0 = Vf(Port(29) | SubPort(0) | Vlan(114))
	eth_29_1 = Vf(Port(29) | SubPort(1) | Vlan(115))
	eth_29_2 = Vf(Port(29) | SubPort(2) | Vlan(116))
	eth_29_3 = Vf(Port(29) | SubPort(3) | Vlan(117))
	eth_30_0 = Vf(Port(30) | SubPort(0) | Vlan(126))
	eth_30_1 = Vf(Port(30) | SubPort(1) | Vlan(127))
	eth_30_2 = Vf(Port(30) | SubPort(2) | Vlan(128))
	eth_30_3 = Vf(Port(30) | SubPort(3) | Vlan(129))
	eth_31_0 = Vf(Port(31) | SubPort(0) | Vlan(122))
	eth_31_1 = Vf(Port(31) | SubPort(1) | Vlan(123))
	eth_31_2 = Vf(Port(31) | SubPort(2) | Vlan(124))
	eth_31_3 = Vf(Port(31) | SubPort(3) | Vlan(125))
)

func mksriovs() error {
	var porto uint
	if ver, err := deviceVersion(); err != nil {
		return err
	} else if ver > 0 {
		porto = 1
	}
	return Mksriovs(porto, []Vf{
		// pf0
		eth_0_0,
		eth_1_0,
		eth_2_0,
		eth_3_0,
		eth_4_0,
		eth_5_0,
		eth_6_0,
		eth_7_0,
		eth_8_0,
		eth_9_0,
		eth_10_0,
		eth_11_0,
		eth_12_0,
		eth_13_0,
		eth_14_0,
		eth_15_0,
		eth_0_1,
		eth_1_1,
		eth_2_1,
		eth_3_1,
		eth_4_1,
		eth_5_1,
		eth_6_1,
		eth_7_1,
		eth_8_1,
		eth_9_1,
		eth_10_1,
		eth_11_1,
		eth_12_1,
		eth_13_1,
		eth_14_1,
		eth_15_1,
		eth_0_2,
		eth_1_2,
		eth_2_2,
		eth_3_2,
		eth_4_2,
		eth_5_2,
		eth_6_2,
		eth_7_2,
		eth_8_2,
		eth_9_2,
		eth_10_2,
		eth_11_2,
		eth_12_2,
		eth_13_2,
		eth_14_2,
		eth_15_2,
		eth_0_3,
		eth_1_3,
		eth_2_3,
		eth_3_3,
		eth_4_3,
		eth_5_3,
		eth_6_3,
		eth_7_3,
		eth_8_3,
		eth_9_3,
		eth_10_3,
		eth_11_3,
		eth_12_3,
		eth_13_3,
		eth_14_3,
		eth_15_3,
	}, []Vf{
		// pf1
		eth_16_0,
		eth_17_0,
		eth_18_0,
		eth_19_0,
		eth_20_0,
		eth_21_0,
		eth_22_0,
		eth_23_0,
		eth_24_0,
		eth_25_0,
		eth_26_0,
		eth_27_0,
		eth_28_0,
		eth_29_0,
		eth_30_0,
		eth_31_0,
		eth_16_1,
		eth_17_1,
		eth_18_1,
		eth_19_1,
		eth_20_1,
		eth_21_1,
		eth_22_1,
		eth_23_1,
		eth_24_1,
		eth_25_1,
		eth_26_1,
		eth_27_1,
		eth_28_1,
		eth_29_1,
		eth_30_1,
		eth_31_1,
		eth_16_2,
		eth_17_2,
		eth_18_2,
		eth_19_2,
		eth_20_2,
		eth_21_2,
		eth_22_2,
		eth_23_2,
		eth_24_2,
		eth_25_2,
		eth_26_2,
		eth_27_2,
		eth_28_2,
		eth_29_2,
		eth_30_2,
		eth_31_2,
		eth_16_3,
		eth_17_3,
		eth_18_3,
		eth_19_3,
		eth_20_3,
		eth_21_3,
		eth_22_3,
		eth_23_3,
		eth_24_3,
		eth_25_3,
		eth_26_3,
		eth_27_3,
		eth_28_3,
		eth_29_3,
		eth_30_3,
		eth_31_3,
	})
}
