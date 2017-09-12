#!/bin/bash
set -x -e

# $1 = ifc_name
# $2 = x.y.z.t/24

ip addr del $2 dev $1
ip add sho $1
