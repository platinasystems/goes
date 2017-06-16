## Using Linux Network Namespaces with Platina interfaces

### Manual Example 
This example uses the `ip` command to manually move interfaces eth-30-0 and eth-31-0 into their own containers.

1. Connect a cable between eth-30-0 and eth-31-0.
2. Create 2 network namespace call net1 and net2:
```
ip netns add net1
ip netns add net2
```   
3. Move eth-30-0 to net1 and eth-31-0 to net2:
```
ip link set eth-30-0 netns net1
ip link set eth-31-0 netns net2
```
After the interfaces have been moved to a different namespace, we can see that they no longer exist in the default namespace:
```sh
# ip link show eth-30-0
Device "eth-30-0" does not exist.

# ip link show eth-31-0
Device "eth-31-0" does not exist.
```
4. Set a different nsid for each namespace. The id just needs to be a number that's currently not used (use `ip netns list-id` to see what are already used). In this example we'll use 10 and 11. These ids are used by the kernel to send netlink messages for a specific namespace.
```
ip netns set net1 10
ip netns set net2 11
```  
5. Bring up the link and add an address to each interface in net1 and net2. To run a command in a namespace we can use the 
`ip netns exec <namespace> <command>`.
```
ip netns exec net1 ip link set up eth-30-0
ip netns exec net1 ip addr add 10.1.0.1/24 dev eth-30-0
```
```
ip netns exec net2 ip link set up eth-31-0
ip netns exec net2 ip addr add 10.1.0.2/24 dev eth-31-0
```   
6. Verify that the interface is in the namespace, up and with the address configured on it.
```sh
# ip netns exec net1 ip add show
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
8978: eth-30-0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP group default qlen 1000
    link/ether 02:46:8a:00:02:c1 brd ff:ff:ff:ff:ff:ff
    inet 10.1.0.1/24 scope global eth-30-0
       valid_lft forever preferred_lft forever
    inet6 fe80::46:8aff:fe00:2c1/64 scope link 
       valid_lft forever preferred_lft forever
```
```sh
# ip netns exec net2 ip add show
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
8979: eth-31-0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP group default qlen 1000
    link/ether 02:46:8a:00:02:c2 brd ff:ff:ff:ff:ff:ff
    inet 10.1.0.2/24 scope global eth-31-0
       valid_lft forever preferred_lft forever
    inet6 fe80::46:8aff:fe00:2c2/64 scope link 
       valid_lft forever preferred_lft forever
```
7. Now we should be able to ping from net1 to net2:
```sh
# ip netns exec net1 ping 10.1.0.2
PING 10.1.0.2 (10.1.0.2) 56(84) bytes of data.
64 bytes from 10.1.0.2: icmp_seq=1 ttl=64 time=0.132 ms
64 bytes from 10.1.0.2: icmp_seq=2 ttl=64 time=0.102 ms
64 bytes from 10.1.0.2: icmp_seq=3 ttl=64 time=0.082 ms
^C
--- 10.1.0.2 ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 2034ms
rtt min/avg/max/mdev = 0.082/0.105/0.132/0.022 ms
```
8. Looking at goes fib we can see that each namespace has it's own table id:
```sh
root@invader1:~# goes vnet show ip fib
 Table                   Destination           Adjacency
     0                  10.15.0.0/24       3: glean eth-0-0
     0                  10.15.0.1/32       4: local eth-0-0
                                        tx pipe packets                3
                                        tx pipe bytes                306
     0                  10.15.0.4/32       5: rewrite eth-0-0 IP4: 02:46:8a:00:02:a3 -> 02:46:8a:00:01:97
     2                   10.1.0.3/32      13: rewrite eth-30-0 IP4: 02:46:8a:00:02:c1 -> 02:46:8a:00:02:c2
     3                   10.1.0.2/32      12: rewrite eth-31-0 IP4: 02:46:8a:00:02:c2 -> 02:46:8a:00:02:c1
     4                   10.1.0.0/24       6: glean eth-30-0
                                        tx pipe packets               10
                                        tx pipe bytes               1020
     4                   10.1.0.2/32      15: rewrite eth-30-0 IP4: 02:46:8a:00:02:c1 -> 02:46:8a:00:02:c2
     5                   10.1.0.0/24      10: glean eth-31-0
                                        tx pipe packets                8
                                        tx pipe bytes                816
     5                   10.1.0.1/32      14: rewrite eth-31-0 IP4: 02:46:8a:00:02:c2 -> 02:46:8a:00:02:c1
     5                   10.1.0.3/32      11: local eth-31-0
                                        tx pipe packets                2
                                        tx pipe bytes                204
```
9. To delete the namespaces we can do them individually or delete them all:
```sh
# ip netns list
net2 (id: 11)
net1 (id: 10)

# ip netns del net1
# ip -all netns del

# ip netns list
# 
```
### Script Example
The `ns_move.sh` can also be used to to accomplish the same thing:
```sh
# ./ns_move.sh up net1 eth-30-0 10.1.0.2/24
# ./ns_move.sh up net2 eth-31-0 10.1.0.3/24
 
# ip netns exec net1 ip addr sh
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host 
       valid_lft forever preferred_lft forever
8978: eth-30-0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP group default qlen 1000
    link/ether 02:46:8a:00:02:c1 brd ff:ff:ff:ff:ff:ff
    inet 10.1.0.2/24 scope global eth-30-0
       valid_lft forever preferred_lft forever
    inet6 fe80::46:8aff:fe00:2c1/64 scope link 
       valid_lft forever preferred_lft forever

# ip netns exec net1 ping 10.1.0.3
PING 10.1.0.3 (10.1.0.3) 56(84) bytes of data.
64 bytes from 10.1.0.3: icmp_seq=1 ttl=64 time=0.259 ms
64 bytes from 10.1.0.3: icmp_seq=2 ttl=64 time=0.121 ms
^C
--- 10.1.0.3 ping statistics ---
2 packets transmitted, 2 received, 0% packet loss, time 1026ms
rtt min/avg/max/mdev = 0.121/0.190/0.259/0.069 ms

```
To bring down the namespaces:
```sh
# ./ns_move.sh down net1 eth-30-0 
# ./ns_move.sh down net2 eth-31-0 
```
After that the interfaces move back to the default namespace:
```sh
# ip add show eth-30-0
8978: eth-30-0: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN group default qlen 1000
    link/ether 02:46:8a:00:02:c1 brd ff:ff:ff:ff:ff:ff

# ip add show eth-31-0
8979: eth-31-0: <BROADCAST,MULTICAST> mtu 1500 qdisc noop state DOWN group default qlen 1000
    link/ether 02:46:8a:00:02:c2 brd ff:ff:ff:ff:ff:ff
```
