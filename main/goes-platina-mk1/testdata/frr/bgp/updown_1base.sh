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

case $1 in
    "up")
	docker-compose -f docker-compose_1base.yaml up -d
	ip link add dummy0 type dummy 2> /dev/null
	ip link add dummy1 type dummy 2> /dev/null
	ip link add dummy2 type dummy 2> /dev/null
	ip link add dummy3 type dummy 2> /dev/null

	# need to force mtu as 9k default doesn't work
	ip link set mtu 1500 dev eth-23-1
	$D_MOVE up R1 eth-23-1 192.168.12.1/24
	ip link set mtu 1500 dev eth-6-1	
	$D_MOVE up R1 eth-6-1 192.168.14.1/24
	$D_MOVE up R1 dummy0 192.168.1.1/32
	# extra dummy1 for route injection testing
	docker exec R1 ip link add dummy1 type dummy
	docker exec R1 ip link set dev dummy1 up

	ip link set mtu 1500 dev eth-24-1
	$D_MOVE up R2 eth-24-1 192.168.12.2/24
	ip link set mtu 1500 dev eth-16-1	
	$D_MOVE up R2 eth-16-1 192.168.23.2/24
	$D_MOVE up R2 dummy1 192.168.2.2/32

	ip link set mtu 1500 dev eth-32-1
	$D_MOVE up R3 eth-32-1 192.168.34.3/24
	ip link set mtu 1500 dev eth-15-1	
	$D_MOVE up R3 eth-15-1 192.168.23.3/24
	$D_MOVE up R3 dummy2 192.168.3.3/32

	ip link set mtu 1500 dev eth-31-1	
	$D_MOVE up R4 eth-31-1 192.168.34.4/24
	ip link set mtu 1500 dev eth-5-1	
	$D_MOVE up R4 eth-5-1 192.168.14.4/24
	$D_MOVE up R4 dummy3 192.168.4.4/32
	;;
    "down")
	$D_MOVE down R1 eth-23-1
	$D_MOVE down R1 eth-6-1
	$D_MOVE down R1 dummy0

	$D_MOVE down R2 eth-24-1
	$D_MOVE down R2 eth-16-1
	$D_MOVE down R2 dummy1

	$D_MOVE down R3 eth-32-1
	$D_MOVE down R3 eth-15-1
	$D_MOVE down R3 dummy2

	$D_MOVE down R4 eth-31-1
	$D_MOVE down R4 eth-5-1
	$D_MOVE down R4 dummy3

	docker-compose -f docker-compose_1base.yaml down
	
	for ns in R1 R2 R3 R4; do
	   ip netn del $ns
	done
	chown -R $USER:$USER volumes
	;;
    *)
	echo "Unknown action"
	;;
esac
