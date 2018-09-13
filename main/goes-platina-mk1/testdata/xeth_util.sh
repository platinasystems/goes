#!/bin/bash

xeth_driver="platina-mk1"
xeth_type="xeth"

xeth_all()
{
    for i in $(ls -1 /sys/class/net); do
        if [ $i == "lo" ]; then
            continue
        fi
        if $(ethtool -i $i | egrep -q -e "driver: xeth" -e VLAN); then
            echo $i
        fi
    done
}

xeth_range()
{
    start=$1
    shift
    stop=$1
    shift

    for i in $(seq $start $stop); do
        echo -n xeth$i" "
    done
}

eth_range()
{
    start=$1
    shift
    stop=$1
    shift

    for i in $(seq $start $stop); do
        echo -n eth-$i-0" "
    done
}

xeth_up()
{
    for i in $xeth_list; do
        ip link set $i up
    done
}

xeth_down()
{
    for i in $xeth_list; do
        ip link set $i down
    done
}

xeth_carrier()
{
    for i in $xeth_list; do
        goes ip link set $i +carrier
    done
}

xeth_no_carrier()
{
    for i in $xeth_list; do
        goes ip link set $i -carrier
    done
}

xeth_flap()
{
    xeth_down $xeth_list
    xeth_up $xeth_list
}

xeth_add()
{
    for i in $xeth_list; do
        ip link add $i type ${xeth_type}
        ip link set $i up
        ethtool -s $i speed 100000 autoneg off
    done
}

xeth_netport_add()
{
    xeth_list=$(grep -o " .eth.*" netport.yaml)
    xeth_add
    xeth_flap $xeth_list
}

xeth_del()
{
    for i in $xeth_list; do
        ip link del $i
    done
}

xeth_br_add()
{
    vid=$1
    shift
    ip link add xethbr.$vid type ${xeth_type}
}

xeth_br_del()
{
    vid=$1
    shift
    ip link del xethbr.$vid type ${xeth_type}
}

xeth_brm_add()
{
    vid=$1
    shift
    taguntag=$1
    shift
    for i in $xeth_list; do
        ip link add $i.$vid$taguntag type ${xeth_type}
    done
}

xeth_brm_del()
{
    vid=$1
    shift
    taguntag=$1
    shift
    for i in $xeth_list; do
        ip link del $i.$vid$taguntag type ${xeth_type}
    done
}

xeth_br_show()
{
    vid=$1
    shift
    ip link | egrep eth.$vid
    ip link | egrep eth[0-9]+.$vid
}

xeth_show()
{
    for i in $xeth_list; do
        #ip link show $i
        ip addr show dev $i
    done
}

xeth_echo()
{
    for i in $xeth_list; do
        echo -n $i" "
    done
    echo
}

xeth_isup()
{
    xeth_show $xeth_list | grep -i state.up | wc -l
}

xeth_stat()
{
    for i in $xeth_list; do
        echo $i
        ethtool -S $i
    done
}

xeth_to_netns()
{
    netns=$1
    shift

    for i in $xeth_list; do
        ip link set $i netns $netns
    done
    ip netns exec $netns ./xeth_util.sh flap
    ip netns exec $netns ./xeth_util.sh show
}

xeth_netns_del()
{
    for netns in $(ip netns); do
      for i in $(ip netns exec $netns ./xeth_util.sh echo); do
        ip netns exec $netns ip link set $i netns 1
      done
      ip netns del $netns
    done
}

xeth_netns_show()
{
    show_ip=false
    show_route=false
    show_vrf=false

    if [ "$1" == "ip" ]; then
      show_ip=true
      shift
    fi
    if [ "$1" == "route" ]; then
      show_route=true
      shift
    fi
    if [ "$1" == "vrf" ]; then
      show_vrf=true
      shift
    fi
    for netns in $(ip netns | sort -V); do
      echo
      echo "netns "$netns
      if $show_ip; then
        ip netns exec $netns ./xeth_util.sh show | grep -e 'inet '|sed -e "s/inet \(.*\) scope global \(.*\)/\2\t\1/"
      else
        ip netns exec $netns ./xeth_util.sh show
      fi
      if $show_route; then
        ip netns exec $netns ip route
      fi
    done
    if $show_vrf; then
      echo ---
      echo vrf
      echo ---
      goes vnet show fe1 l3 |grep -o " intf.xeth.*vrf[^ ]* " |sort -uV|sed -e "s/intf.//; s/vrf.//; s/\/.*//" |grep -v 0$
    fi
}

xeth_netns_echo()
{
    for netns in $(ip netns | sort -V); do
      echo -n "netns "$netns": "
      ip netns exec $netns ./xeth_util.sh echo
    done
}

xeth_netns_flap()
{
    for netns in $(ip netns); do
      echo "netns "$netns
      ip netns exec $netns ./xeth_util.sh flap
    done
}

xeth_netns_carrier()
{
    for netns in $(ip netns); do
      echo "netns "$netns
      ip netns exec $netns ./xeth_util.sh carrier
    done
}

range="all"
if [ $# -gt 0 ]; then
    range=$1
fi

if [ $range == "xeth_range" ]; then
    shift
    start=$1
    shift
    stop=$1
    shift
    xeth_list=$(xeth_range $start $stop)

elif [ $range == "eth_range" ]; then
    shift
    start=$1
    shift
    stop=$1
    shift
    xeth_list=$(eth_range $start $stop)

else
    xeth_list="$(xeth_all | sort -V)"
fi

cmd="help"
if [ $# -gt 0 ]; then
    cmd=$1
    shift
fi

# echo range = $xeth_list
# echo command = $cmd

if [ $cmd == "show" ]; then
    xeth_show $xeth_list
elif [ $cmd == "showup" ]; then
    xeth_show $xeth_list | grep -i state.up
elif [ $cmd == "echo" ]; then
    xeth_echo $xeth_list
elif [ $cmd == "reset" ]; then
    # FIXME: also remove vlan interfaces
    xeth_netns_del
    rmmod ${xeth_driver}
    modprobe ${xeth_driver}
elif [ $cmd == "test_init" ]; then
    rmmod ${xeth_driver}
    modprobe ${xeth_driver}
elif [ $cmd == "add" ]; then
    xeth_add $xeth_list
elif [ $cmd == "br_add" ]; then
    xeth_br_add $1
elif [ $cmd == "br_del" ]; then
    xeth_br_del $1
elif [ $cmd == "brm_add" ]; then
    xeth_brm_add $1 $2 $xeth_list
elif [ $cmd == "brm_del" ]; then
    xeth_brm_del $1 $2 $xeth_list
elif [ $cmd == "br_show" ]; then
    xeth_br_show $1
elif [ $cmd == "up" ]; then
    xeth_up $xeth_list
elif [ $cmd == "carrier" ]; then
    xeth_carrier $xeth_list
elif [ $cmd == "down" ]; then
    xeth_down $xeth_list
elif [ $cmd == "flap" ]; then
    xeth_flap $xeth_list
elif [ $cmd == "isup" ]; then
    xeth_isup
elif [ $cmd == "stat" ]; then
    xeth_stat $xeth_list | grep -v " 0$"
elif [ $cmd == "to_netns" ]; then
    xeth_to_netns $1 $xeth_list
elif [ $cmd == "netns_del" ]; then
    xeth_netns_del
elif [ $cmd == "netns_show" ]; then
    xeth_netns_show $*
elif [ $cmd == "netns_echo" ]; then
    echo "default: "$(xeth_echo)
    xeth_netns_echo
elif [ $cmd == "netns_flap" ]; then
    xeth_netns_flap
elif [ $cmd == "netns_carrier" ]; then
    xeth_netns_carrier
else
    # help
    grep range.*[t]hen $0 | grep -o \".*\"
    grep cmd.*[t]hen $0 | grep -o \".*\"
fi
