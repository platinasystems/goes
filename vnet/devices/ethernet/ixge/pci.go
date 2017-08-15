// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ixge

import (
	"github.com/platinasystems/go/elib/hw/pci"

	"fmt"
)

// PCI dev IDs
const (
	dev_id_82598                 = 0x10b6
	dev_id_82598_bx              = 0x1508
	dev_id_82598af_dual_port     = 0x10c6
	dev_id_82598af_single_port   = 0x10c7
	dev_id_82598eb_sfp_lom       = 0x10db
	dev_id_82598at               = 0x10c8
	dev_id_82598at2              = 0x150b
	dev_id_82598eb_cx4           = 0x10dd
	dev_id_82598_cx4_dual_port   = 0x10ec
	dev_id_82598_da_dual_port    = 0x10f1
	dev_id_82598_sr_dual_port_em = 0x10e1
	dev_id_82598eb_xf_lr         = 0x10f4
	dev_id_82599_kx4             = 0x10f7
	dev_id_82599_kx4_mezz        = 0x1514
	dev_id_82599_kr              = 0x1517
	dev_id_82599_t3_lom          = 0x151c
	dev_id_82599_cx4             = 0x10f9
	dev_id_82599_sfp             = 0x10fb
	sub_dev_id_82599_sfp         = 0x11a9
	sub_dev_id_82599_sfp_wol0    = 0x1071
	sub_dev_id_82599_rndc        = 0x1f72
	sub_dev_id_82599_560flr      = 0x17d0
	sub_dev_id_82599_sp_560flr   = 0x211b
	sub_dev_id_82599_ecna_dp     = 0x0470
	sub_dev_id_82599_lom_sfp     = 0x8976
	dev_id_82599_backplane_fcoe  = 0x152a
	dev_id_82599_sfp_fcoe        = 0x1529
	dev_id_82599_sfp_em          = 0x1507
	dev_id_82599_sfp_sf2         = 0x154d
	dev_id_82599en_sfp           = 0x1557
	sub_dev_id_82599en_sfp_ocp1  = 0x0001
	dev_id_82599_xaui_lom        = 0x10fc
	dev_id_82599_combo_backplane = 0x10f8
	sub_dev_id_82599_kx4_kr_mezz = 0x000c
	dev_id_82599_ls              = 0x154f
	dev_id_x540t                 = 0x1528
	dev_id_82599_sfp_sf_qp       = 0x154a
	dev_id_82599_qsfp_sf_qp      = 0x1558
	dev_id_x540t1                = 0x1560
	dev_id_x550t                 = 0x1563
	dev_id_x550em_x_kx4          = 0x15aa
	dev_id_x550em_x_kr           = 0x15ab
	dev_id_x550em_x_sfp          = 0x15ac
	dev_id_x550em_x_10g_t        = 0x15ad
	dev_id_x550em_x_1g_t         = 0x15ae
	dev_id_x550_vf_hv            = 0x1564
	dev_id_x550_vf               = 0x1565
	dev_id_x550em_x_vf           = 0x15a8
	dev_id_x550em_x_vf_hv        = 0x15a9
)

type dev_id pci.VendorDeviceID

func (d *dev) get_dev_id() dev_id { return dev_id(d.pci_dev.DeviceID()) }

func (d dev_id) String() (v string) {
	var ok bool
	if v, ok = dev_id_names[d]; !ok {
		v = fmt.Sprintf("unknown %04x", uint(d))
	}
	return
}

var dev_id_names = map[dev_id]string{
	dev_id_82598:                 "82598",
	dev_id_82598_bx:              "82598_BX",
	dev_id_82598af_dual_port:     "82598AF_DUAL_PORT",
	dev_id_82598af_single_port:   "82598AF_SINGLE_PORT",
	dev_id_82598eb_sfp_lom:       "82598EB_SFP_LOM",
	dev_id_82598at:               "82598AT",
	dev_id_82598at2:              "82598AT2",
	dev_id_82598eb_cx4:           "82598EB_CX4",
	dev_id_82598_cx4_dual_port:   "82598_CX4_DUAL_PORT",
	dev_id_82598_da_dual_port:    "82598_DA_DUAL_PORT",
	dev_id_82598_sr_dual_port_em: "82598_SR_DUAL_PORT_EM",
	dev_id_82598eb_xf_lr:         "82598EB_XF_LR",
	dev_id_82599_kx4:             "82599_KX4",
	dev_id_82599_kx4_mezz:        "82599_KX4_MEZZ",
	dev_id_82599_kr:              "82599_KR",
	dev_id_82599_t3_lom:          "82599_T3_LOM",
	dev_id_82599_cx4:             "82599_CX4",
	dev_id_82599_sfp:             "82599_SFP",
	dev_id_82599_backplane_fcoe:  "82599_BACKPLANE_FCOE",
	dev_id_82599_sfp_fcoe:        "82599_SFP_FCOE",
	dev_id_82599_sfp_em:          "82599_SFP_EM",
	dev_id_82599_sfp_sf2:         "82599_SFP_SF2",
	dev_id_82599en_sfp:           "82599EN_SFP",
	dev_id_82599_xaui_lom:        "82599_XAUI_LOM",
	dev_id_82599_combo_backplane: "82599_COMBO_BACKPLANE",
	dev_id_82599_ls:              "82599_LS",
	dev_id_x540t:                 "X540T",
	dev_id_82599_sfp_sf_qp:       "82599_SFP_SF_QP",
	dev_id_82599_qsfp_sf_qp:      "82599_QSFP_SF_QP",
	dev_id_x540t1:                "X540T1",
	dev_id_x550t:                 "X550T",
	dev_id_x550em_x_kx4:          "X550EM_X_KX4",
	dev_id_x550em_x_kr:           "X550EM_X_KR",
	dev_id_x550em_x_sfp:          "X550EM_X_SFP",
	dev_id_x550em_x_10g_t:        "X550EM_X_10G_T",
	dev_id_x550em_x_1g_t:         "X550EM_X_1G_T",
	dev_id_x550_vf_hv:            "X550_VF_HV",
	dev_id_x550_vf:               "X550_VF",
	dev_id_x550em_x_vf:           "X550EM_X_VF",
	dev_id_x550em_x_vf_hv:        "X550EM_X_VF_HV",
}
