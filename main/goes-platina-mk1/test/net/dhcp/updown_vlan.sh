#!/bin/bash

if ! [ $(id -u) = 0 ]; then
   echo "You must run this script as root/sudo."
   exit 1
fi
USER=$SUDO_USER

if [ -z "$1" ]; then
    echo "Usages: $0 up|down"
    exit 1
fi

D_MOVE=../docker_move.sh

# A loopback cable is connected between side A and B.
# All vlans go over this cable
SIDE_A=eth-4-0
SIDE_B=eth-5-0

case $1 in
    "up")
	docker-compose up -d

	ip link set up $SIDE_A
	ip link set up $SIDE_B
	sleep 1

	for intf in $SIDE_A $SIDE_B; do
	    for vlan in 10 ; do
		ip link add link $intf name $intf.$vlan type vlan id $vlan
		ip link set up $intf.$vlan
	    done	    
	done

	$D_MOVE up R1 $SIDE_A.10 
	$D_MOVE up R2 $SIDE_B.10 192.168.120.10/24
	;;
    "down")
	$D_MOVE down R1 $SIDE_A.10 
	$D_MOVE down R2 $SIDE_B.10
	
	for intf in $SIDE_A $SIDE_B; do
	    for vlan in 10 ; do
		ip link set down $intf.$vlan
		ip link del $intf.$vlan 
	    done	    
	done
		
	docker-compose down
	
	for ns in R1 R2 ; do
	   ip netn del $ns
	done
	;;
    *)
	echo "Unknown action"
	;;
esac
