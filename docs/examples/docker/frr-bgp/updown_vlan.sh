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
	ip link add dummy0 type dummy 2> /dev/null
	ip link add dummy1 type dummy 2> /dev/null
	ip link add dummy2 type dummy 2> /dev/null
	ip link add dummy3 type dummy 2> /dev/null

	ip link set up $SIDE_A
	ip link set up $SIDE_B
	sleep 1

	for intf in $SIDE_A $SIDE_B; do
	    for vlan in 10 20 30 40 50 60; do
		ip link add link $intf name $intf.$vlan type vlan id $vlan
		ip link set up $intf.$vlan
	    done	    
	done

	ip link set mtu 1500 dev $SIDE_A
	ip link set mtu 1500 dev $SIDE_B	
	$D_MOVE up R1 $SIDE_A.10 192.168.120.5/24
	$D_MOVE up R1 $SIDE_B.40 192.168.150.5/24
	$D_MOVE up R1 $SIDE_A.50 192.168.50.5/24
	$D_MOVE up R1 dummy0 192.168.1.5/32

	$D_MOVE up R2 $SIDE_B.10 192.168.120.10/24
	$D_MOVE up R2 $SIDE_A.20 192.168.222.10/24
	$D_MOVE up R2 $SIDE_A.60 192.168.60.10/24
	$D_MOVE up R2 dummy1 192.168.1.10/32

	$D_MOVE up R3 $SIDE_A.30 192.168.111.2/24
	$D_MOVE up R3 $SIDE_B.20 192.168.222.2/24
	$D_MOVE up R3 $SIDE_B.50 192.168.50.2/24	
	$D_MOVE up R3 dummy2 192.168.2.2/32

	$D_MOVE up R4 $SIDE_B.30 192.168.111.4/24
	$D_MOVE up R4 $SIDE_A.40 192.168.150.4/24
	$D_MOVE up R4 $SIDE_B.60 192.168.60.4/24	
	$D_MOVE up R4 dummy3 192.168.2.4/32
	;;
    "down")
	$D_MOVE down R1 $SIDE_A.10 
	$D_MOVE down R1 $SIDE_B.40 
	$D_MOVE down R1 $SIDE_A.50 
	$D_MOVE down R1 dummy0

	$D_MOVE down R2 $SIDE_B.10
	$D_MOVE down R2 $SIDE_A.20
	$D_MOVE down R2 $SIDE_A.60
	$D_MOVE down R2 dummy1

	$D_MOVE down R3 $SIDE_A.30
	$D_MOVE down R3 $SIDE_B.20
	$D_MOVE down R3 $SIDE_B.50	
	$D_MOVE down R3 dummy2

	$D_MOVE down R4 $SIDE_B.30
	$D_MOVE down R4 $SIDE_A.40
	$D_MOVE down R4 $SIDE_B.60	
	$D_MOVE down R4 dummy3

	for intf in $SIDE_A $SIDE_B; do
	    for vlan in 10 20 30 40 50 60; do
		ip link set down $intf.$vlan
		ip link del $intf.$vlan 
	    done	    
	done
		
	docker-compose down
	
	for ns in R1 R2 R3 R4; do
	   ip netn del $ns
	done
	chown -R $USER:$USER volumes
	;;
    *)
	echo "Unknown action"
	;;
esac
