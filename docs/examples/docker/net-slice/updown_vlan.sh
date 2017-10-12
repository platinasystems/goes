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
SIDE_A=eth-30-0
SIDE_B=eth-31-0

case $1 in
    "up")
	docker-compose up -d

	ip link set up $SIDE_A
	ip link set up $SIDE_B
	sleep 1

	for intf in $SIDE_A $SIDE_B; do
	    for vlan in 10 20 30 40 50 60 ; do
		ip link add link $intf name $intf.$vlan type vlan id $vlan
		ip link set up $intf.$vlan
	    done	    

	done

	$D_MOVE up CA-1 $SIDE_A.10 10.1.0.1/24	
	$D_MOVE up RA-1 $SIDE_B.10 10.1.0.2/24
	$D_MOVE up RA-1 $SIDE_A.20 10.2.0.2/24
	$D_MOVE up RA-2 $SIDE_B.20 10.2.0.3/24
	$D_MOVE up RA-2 $SIDE_A.30 10.3.0.3/24
	$D_MOVE up CA-2 $SIDE_B.30 10.3.0.4/24

	$D_MOVE up CB-1 $SIDE_A.40 10.1.0.1/24	
	$D_MOVE up RB-1 $SIDE_B.40 10.1.0.2/24
	$D_MOVE up RB-1 $SIDE_A.50 10.2.0.2/24
	$D_MOVE up RB-2 $SIDE_B.50 10.2.0.3/24
	$D_MOVE up RB-2 $SIDE_A.60 10.3.0.3/24
	$D_MOVE up CB-2 $SIDE_B.60 10.3.0.4/24			
	;;
    "down")
	$D_MOVE down CA-1 $SIDE_A.10 
	$D_MOVE down RA-1 $SIDE_B.10 
	$D_MOVE down RA-1 $SIDE_A.20 
	$D_MOVE down RA-2 $SIDE_B.20 
	$D_MOVE down RA-2 $SIDE_A.30 
	$D_MOVE down CA-2 $SIDE_B.30 

	$D_MOVE down CB-1 $SIDE_A.40 
	$D_MOVE down RB-1 $SIDE_B.40 
	$D_MOVE down RB-1 $SIDE_A.50 
	$D_MOVE down RB-2 $SIDE_B.50 
	$D_MOVE down RB-2 $SIDE_A.60 
	$D_MOVE down CB-2 $SIDE_B.60 

	for intf in $SIDE_A $SIDE_B; do
	    for vlan in 10 20 30 40 50 60 ; do
		ip link set down $intf.$vlan
		ip link del $intf.$vlan 
	    done	    
	done
		
	docker-compose down
	
	for ns in CA-1 RA-1 RA-2 CA-2 CB-1 RB-1 RB-2 CB-2; do
	   ip netn del $ns
	done
	chown -R $USER:$USER volumes
	;;
    *)
	echo "Unknown action"
	;;
esac
