#!/bin/bash

#
# Copyright 2015-2017 Platina Systems, Inc. All rights reserved.
# Use of this source code is governed by the GPL-2 license described in the
# LICENSE file.
#
# Script to find the pid of a docker container
#

if ! [ $(id -u) = 0 ]; then
   echo "You must run this script as root/sudo."
   exit 1
fi

if [ $# -lt 1 ]; then
    echo "Usage: $0 <container_name>"
    echo "eg. $0 R1"
    exit 1
fi

dc=$1

dc_pid=$(docker inspect -f '{{.State.Pid}}' $dc)
if [ -z "$dc_pid" ]; then
  echo "Docker container [$dc] not found."
  exit 1
fi

echo $dc_pid

exit 0
