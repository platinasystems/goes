#!/bin/bash
set -x -e

# $1 = ifc_name
# $2 = x.y.z.t/24

# If address already exists, then script will error,
# so best to delete address with other task before
# this task.

# "brd +" automatically sets the broadcast address.
ip addr add $2 brd + dev $1
ip link set $1 up
ip add sho $1
