// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package diag

const i2cmon_min, i2cmon_max = true, true
const irq_l_on_min, irq_l_on_max = 0, 0
const irq_l_off_min, irq_l_off_max = 1, 1
const active_high_off_min, active_high_off_max = false, false
const active_high_on_min, active_high_on_max = true, true
const active_low_off_min, active_low_off_max = true, true
const active_low_on_min, active_low_on_max = false, false
const ping_noresponse_min, ping_noresponse_max = false, false
const ping_response_min, ping_response_max = true, true
const i2cping_noresponse_min, i2cping_noresponse_max = false, false
const i2cping_response_min, i2cping_response_max = true, true
const vmon_1v0_tha_min, vmon_1v0_tha_max = 0.983, 1.087
const vmon_1v0_thc_min, vmon_1v0_thc_max = 0.950, 1.050
const vmon_1v2_ethx_min, vmon_1v2_ethx_max = 1.114, 1.260
const vmon_1v25_sys_min, vmon_1v25_sys_max = 1.187, 1.312
const vmon_1v8_sys_min, vmon_1v8_sys_max = 1.71, 1.89
const vmon_3v3_bmc_min, vmon_3v3_bmc_max = 3.135, 3.465
const vmon_3v3_sb_min, vmon_3v3_sb_max = 3.135, 3.465
const vmon_3v3_sys_min, vmon_3v3_sys_max = 3.135, 3.465
const vmon_3v8_bmc_min, vmon_3v8_bmc_max = 3.61, 3.99
const vmon_5v0_sb_min, vmon_5v0_sb_max = 4.75, 5.25
const fanspeedlow_min, fanspeedlow_max = 5000, 9000
const fanspeedmed_min, fanspeedmed_max = 10000, 14000
const fanspeedhigh_min, fanspeedhigh_max = 19000, 30000
const tmon_bmc_cpu_min, tmon_bmc_cpu_max = 20.00, 70.00
const tmon_fan_rear_min, tmon_fan_rear_max = 20.00, 80.00
const tmon_fan_front_min, tmon_fan_front_max = 20.00, 80.00
