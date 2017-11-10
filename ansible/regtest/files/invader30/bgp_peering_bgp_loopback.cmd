ifconfig lo 192.168.30.1 netmask 255.255.255.0
route add -net 192.168.31.0 netmask 255.255.255.0 gw 10.0.19.31
route add -net 192.168.32.0 netmask 255.255.255.0 gw 10.0.3.32
