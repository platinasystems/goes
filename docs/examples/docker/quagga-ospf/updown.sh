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
F_NS=../find_ns.sh

case $1 in
    "up")
	docker-compose up -d
	ip link add dummy0 type dummy 2> /dev/null
	ip link add dummy1 type dummy 2> /dev/null
	ip link add dummy2 type dummy 2> /dev/null
	ip link add dummy3 type dummy 2> /dev/null

	$D_MOVE up R1 eth-25-0 192.168.120.5/24
	$D_MOVE up R1 eth-4-0 192.168.150.5/24
	$D_MOVE up R1 dummy0 192.168.1.5/32

	$D_MOVE up R2 eth-24-0 192.168.120.10/24
	$D_MOVE up R2 eth-14-0 192.168.222.10/24
	$D_MOVE up R2 dummy1 192.168.1.10/32

	$D_MOVE up R3 eth-30-0 192.168.111.2/24
	$D_MOVE up R3 eth-15-0 192.168.222.2/24
	$D_MOVE up R3 dummy2 192.168.2.2/32

	$D_MOVE up R4 eth-31-0 192.168.111.4/24
	$D_MOVE up R4 eth-5-0 192.168.150.4/24
	$D_MOVE up R4 dummy3 192.168.2.4/32
	;;
    "down")
	$D_MOVE down R1 eth-25-0
	$D_MOVE down R1 eth-4-0
	$D_MOVE down R1 dummy0
	R1_NS=$($F_NS R1)

	$D_MOVE down R2 eth-24-0
	$D_MOVE down R2 eth-14-0
	$D_MOVE down R2 dummy1
	R2_NS=$($F_NS R2)	

	$D_MOVE down R3 eth-30-0
	$D_MOVE down R3 eth-15-0
	$D_MOVE down R3 dummy2
	R3_NS=$($F_NS R3)	

	$D_MOVE down R4 eth-31-0
	$D_MOVE down R4 eth-5-0
	$D_MOVE down R4 dummy3
	R4_NS=$($F_NS R4)	
	docker-compose down
	ip netn del $R1_NS
	ip netn del $R2_NS
	ip netn del $R3_NS
	ip netn del $R4_NS	
	chown -R $USER:$USER volumes
	;;
    *)
	echo "Unknown action"
	;;
esac
