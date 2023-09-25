#!/bin/bash

#
# Copyright 2015-2017 Platina Systems, Inc. All rights reserved.
# Use of this source code is governed by the GPL-2 license described in the
# LICENSE file.
#
# Script to create a namespace and move an interface into the namespace
#

if [ "$(id -u)" != "0" ]; then
    echo "Must be run as root"
    exit 1
fi

usage () {
    echo "Usage: $0 up <namespace> <intf> <ip_addr_with_prefix>"
    echo "Usage: $0 down <namespace> <intf>"
    exit 1
}

if [ $# -lt 1 ]; then
    usage
fi	

action=$1

case $action in
    "up")
	if [ $# -ne 4 ]; then
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

ns=$2
intf=$3

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

setup_ns () {
    ns=$1
    intf=$2
    addr_mask=$3

    ip netns add $ns 2> /dev/null
    if [ $? -ne 0 ]; then
	echo "Error: failed to create namespace [$ns]."
	exit 1
    fi
    
    set_nsid $ns
    ip link set $intf netns $ns
    ip netns exec $ns ip link set up lo
    ip netns exec $ns ip add add 127.0.0.1/8 dev lo 2> /dev/null
    ip netns exec $ns ip link set up $intf

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
	ip netns exec $ns ip addr add $addr peer $peer dev $intf
	rc=$?
    else
	ip netns exec $ns ip addr add $addr_mask dev $intf
	rc=$?	
    fi
    if [ $rc -ne 0 ]; then
	echo "Failed to set ip address."
	exit 1
    fi

}

check_ns () {
    ns=$1

    ip netns list-id | grep -q $ns
    if [ $? == 0 ]; then
	echo "Error: namespace [$ns] already in use."
	exit 1
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

up_it () {
    ns=$1
    intf=$2
    addr=$3

    check_ns $ns
    check_intf $intf
    setup_ns $ns $intf $addr
}

kill_processes () {
    pids=$(ip netns pid $1)
    for p in ${pids[@]}; do
	kill $p
	if ( ps -p $p > /dev/null )
	then
	    kill -9 $p
	fi
    done
}

return_intf () {
    ns=$1
    intf=$2

    ip netns exec $ns ip link set $intf netns 1
}

del_ns () {
    ns=$1
    
    p=$(ip netns pid $ns | wc -l)
    if [ "$p" -eq "0" ]; then
	ip netns delete $ns	
    else
	echo "not deleting namespace $ns with process running"
    fi
}

down_it () {
    ns=$1
    intf=$2

    kill_processes $ns

    return_intf $ns $intf

    del_ns $ns
}

case "$1" in
    up)
	up_it $ns $intf $addr
	;;
    
    down)
	down_it $ns $intf
	;;
    
    *)
	usage
esac

if ( ! goes status > /dev/null )
then
    echo "goes not happy"
    exit 1
fi

exit 0
