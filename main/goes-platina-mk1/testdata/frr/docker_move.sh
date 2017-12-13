#!/bin/bash

#
# Copyright 2015-2017 Platina Systems, Inc. All rights reserved.
# Use of this source code is governed by the GPL-2 license described in the
# LICENSE file.
#
# Script to move an interface into a docker container
#

if [ "$(id -u)" != "0" ]; then
    echo "Must be run as root"
    exit 1
fi

usage () {
    echo "Usage: $0 up <docker container> <intf> <ip_addr_with_prefix>"
    echo "Usage: $0 up <docker container> <intf>"
    echo "Usage: $0 down <docker container> <intf>"
    exit 1
}

if [ $# -lt 1 ]; then
    usage
fi	

action=$1

case $action in
    "up")
	if [ $# -ne 3 ] && [ $# -ne 4 ]; then
	    usage
	fi
	addr=$4
	;;
    "down")
	if [ $# -ne 3 ]; then
	    usage
	fi
	;;
    *)
	usage
	;;
esac

dc=$2
intf=$3

find_dc_pid () {
    dc=$1
    
    dc_pid=$(docker inspect -f '{{.State.Pid}}' $dc)
    if [ -z "$dc_pid" ]; then
	echo "Error: docker container [$dc] - pid not found"
	exit 1
    fi
    echo $dc_pid
}

set_nsid () {
    ns=$1

    nsid=1
    ok=0
    # find next available nsid
    while [ "$ok" -eq "0" ]; do
	ip netns set $ns $nsid 2> /dev/null
	if [ "$?" -eq 0 ]; then
	    ok=1
	else
	    ((nsid++))
	fi
    done
}

setup_dc () {
    dc=$1
    intf=$2
    addr_mask=$3

    if [ ! -d "/var/run/netns" ]; then
	mkdir -p /var/run/netns
    fi

    if [ ! -h /var/run/netns/$dc ]; then
	dc_pid=$(find_dc_pid $dc)
	ln -s /proc/$dc_pid/ns/net /var/run/netns/$dc
	#set_nsid $dc    # doesn't seem necessary for docker
    fi

    ip link set $intf netns $dc
    if [ $? -ne 0 ]; then
       echo "Error: set netns failed."
       exit 1
    fi

    ip netns exec $dc ip link set up lo
    ip netns exec $dc ip add add 127.0.0.1/8 dev lo 2> /dev/null
    ip netns exec $dc ip link set up $intf

    if [ -z "$addr_mask" ]; then
	return
    fi

    if [[ "$addr_mask" =~ "/31" ]]; then
	addr=${addr_mask%/*}
	mask=${addr_mask#*/}
	fto=${addr%.*}  # first three octets
	lo=${addr##*.}  # last octet
	rem=$((lo % 2))
	if [ $rem -eq 0 ]; then
	    let peer_oct=lo+1
	else
	    let peer_oct=lo-1
	fi
	peer="$fto.$peer_oct/31"
	ip netns exec $dc ip addr add $addr peer $peer dev $intf
	rc=$?
    else
	ip netns exec $dc ip addr add $addr_mask dev $intf
	rc=$?	
    fi
    if [ $rc -ne 0 ]; then
	echo "Failed to set ip address."
	exit 1
    fi
}

check_dc () {
    dc=$1

    state=$(docker inspect -f '{{.State.Status}}' $dc)
    if [ $? -ne 0 ]; then
	echo "Error: docker container [$dc] not found."
	exit 1
    fi
    if [ "$state" != "running" ]; then
	echo "Error: docker container [$dc] not running - state [$state]"
    fi
}

check_intf () {
    intf=$1

    ip link show $intf &> /dev/null
    if [ $? != 0 ]; then
	echo "Error: interface [$intf] not found in default namespace."
	exit 1
    fi
}

check_intf_dc () {
    dc=$1
    intf=$2

    docker exec -it $dc ip link show $intf &> /dev/null
    if [ $? != 0 ]; then
	echo "Error: interface [$intf] not found in docker container [$dc]."
	exit 1
    fi
}

up_it () {
    dc=$1
    intf=$2
    addr=$3

    check_dc $dc
    check_intf $intf
    setup_dc $dc $intf $addr
}

kill_processes () {
    dc=$1
    dc_pid=$2

    pids=$(ip netns pid $dc)
    for p in ${pids[@]}; do
	if [ "$p" -eq "$dc_pid" ]; then
	    continue   # don't kill the container
	fi
	kill $p
	if ( ps -p $p > /dev/null )
	then
	    kill -9 $p 2> /dev/null
	fi
    done
}

return_intf () {
    dc=$1
    intf=$2

    ip netns exec $dc ip link set down $intf
    ip netns exec $dc ip link set $intf netns 1
}

down_it () {
    dc=$1
    intf=$2

    check_dc $dc
    check_intf_dc $dc $intf    
    
    dc_pid=$(find_dc_pid $dc)
    
    kill_processes $dc $dc_pid

    return_intf $dc $intf
}

case "$1" in
    up)
	up_it $dc $intf $addr
	;;
    
    down)
	down_it $dc $intf
	;;
    
    *)
	usage
esac

if ( ! goes status > /dev/null ); then
    echo "goes not happy"
    exit 1
fi

exit 0
