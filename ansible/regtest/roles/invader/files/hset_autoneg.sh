#!/bin/bash

set -x -e

# Note, pktgen/bcm_shell peer requires:
#   BCM.0> phy <port> AN_X4_CL73_CTLSr PD_KX_EN=0
# eg.
#   BCM.0> phy ce27 AN_X4_CL73_CTLSr PD_KX_EN=0
/usr/bin/goes hset platina vnet.$1.media copper
/usr/bin/goes hset platina vnet.$1.speed auto

