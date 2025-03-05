## deploy
```console
$ ko apply -B --platform=linux/GOARCH -f ./examples/vpn.yaml
2025/03/05 10:21:20 Using base cgr.dev/chainguard/static:latest@sha256:7a6456cc96ecde793b7c8ad9a3ccd5d610d6168a6f64d693ecc2e84f8276c6c6 for github.com/platinasystems/goes/v2
2025/03/05 10:21:21 Building github.com/platinasystems/goes/v2 for linux/arm64
2025/03/05 10:21:21 Loading ko.local/goes:35d83cbc824387604160b522760c80d7314383744ee5bc14d104ff6fcdf4887d
2025/03/05 10:21:21 Loaded ko.local/goes:35d83cbc824387604160b522760c80d7314383744ee5bc14d104ff6fcdf4887d
2025/03/05 10:21:21 Adding tag latest
2025/03/05 10:21:21 Added tag latest
service/vpn created
secret/vpn created
configmap/vpn created
pod/vpn-registry created
pod/vpn-exchange0 created
pod/vpn-hosta created
pod/vpn-hostb created
pod/vpn-hostc created
```

## Check logs
```console
$ kubectl logs vpn-registry
registry.go:205: checkin: id: unavailable
registry.go:205: checkin: id: unavailable
registry.go:205: checkin: id: unavailable
registry.go:528: new service exchange0 @ 10.42.0.20:8003
registry.go:610: exchange0 assigned 0 @ fc00:1234::10
registry.go:532: new quest hostc via 0
registry.go:607: hostc assigned 1 @ fc00:1234::3 via 0
registry.go:532: new quest hostb via 0
registry.go:607: hostb assigned 2 @ fc00:1234::2 via 0
registry.go:532: new quest hosta via 0
registry.go:607: hosta assigned 3 @ fc00:1234::1 via 0
```
```console
kubectl logs vpn-exchange0
rest.go:355: crt file: /etc/goes/vpn/exchange0.pem
rest.go:360: sig file: /etc/goes/vpn/exchange0.pk8
rest.go:380: url file: https://registry.vpn.default.svc.cluster.local:8003
vpn.go:147: link-local address: fe80::be81:2bb4:cf4e:1c99
goresolve.go:54: wait for registry.vpn.default.svc.cluster.local ...
goresolve.go:40: found registry.vpn.default.svc.cluster.local [10.42.0.24] 6.1s
client.go:217: registry.vpn.default.svc.cluster.local [10.42.0.24]
rest.go:285: checkin try 1
client.go:121: assigned id 0 @ fc00:1234::10/128, via 0, vpn fc00:1234::/64
exchange.go:133: whois 2@10.42.0.21:35146
exchange.go:133: whois 3@10.42.0.23:55505
exchange.go:133: whois 3@10.42.0.23:55505
exchange.go:133: whois 2@10.42.0.21:35146
exchange.go:133: whois 1@10.42.0.22:48792
exchange.go:133: whois 1@10.42.0.22:48792
exchange.go:490: tx 2@10.42.0.21:35146 ip6 48 bytes fe80::ab65:d94c:5dcd:21ee <- fe80::be81:2bb4:cf4e:1c99; icmp6 sum ok; router-advertisementhop-limit 2; prefix-information length 64, autonomous-address-configuration, on-link, valid-lifetime infinite, preferred-lifetime infinite, prefix fc00:5678::
exchange.go:490: tx 3@10.42.0.23:55505 ip6 48 bytes fe80::b1d7:552c:c08:2d5 <- fe80::be81:2bb4:cf4e:1c99; icmp6 sum ok; router-advertisementhop-limit 2; prefix-information length 64, autonomous-address-configuration, on-link, valid-lifetime infinite, preferred-lifetime infinite, prefix fc00:5678::
exchange.go:490: tx 1@10.42.0.22:48792 ip6 48 bytes fe80::f65e:7bff:bbde:7b9b <- fe80::be81:2bb4:cf4e:1c99; icmp6 sum ok; router-advertisementhop-limit 2; prefix-information length 64, autonomous-address-configuration, on-link, valid-lifetime infinite, preferred-lifetime infinite, prefix fc00:5678::
exchange.go:490: tx 3@10.42.0.23:55505 ip6 48 bytes fe80::b1d7:552c:c08:2d5 <- fe80::be81:2bb4:cf4e:1c99; icmp6 sum ok; router-advertisementhop-limit 2; prefix-information length 64, autonomous-address-configuration, on-link, valid-lifetime infinite, preferred-lifetime infinite, prefix fc00:5678::
exchange.go:490: tx 2@10.42.0.21:35146 ip6 48 bytes fe80::ab65:d94c:5dcd:21ee <- fe80::be81:2bb4:cf4e:1c99; icmp6 sum ok; router-advertisementhop-limit 2; prefix-information length 64, autonomous-address-configuration, on-link, valid-lifetime infinite, preferred-lifetime infinite, prefix fc00:5678::
exchange.go:490: tx 1@10.42.0.22:48792 ip6 48 bytes fe80::f65e:7bff:bbde:7b9b <- fe80::be81:2bb4:cf4e:1c99; icmp6 sum ok; router-advertisementhop-limit 2; prefix-information length 64, autonomous-address-configuration, on-link, valid-lifetime infinite, preferred-lifetime infinite, prefix fc00:5678::
exchange.go:490: tx 3@10.42.0.23:55505 ip6 48 bytes fe80::b1d7:552c:c08:2d5 <- fe80::be81:2bb4:cf4e:1c99; icmp6 sum ok; router-advertisementhop-limit 2; prefix-information length 64, autonomous-address-configuration, on-link, valid-lifetime infinite, preferred-lifetime infinite, prefix fc00:5678::
exchange.go:490: tx 2@10.42.0.21:35146 ip6 48 bytes fe80::ab65:d94c:5dcd:21ee <- fe80::be81:2bb4:cf4e:1c99; icmp6 sum ok; router-advertisementhop-limit 2; prefix-information length 64, autonomous-address-configuration, on-link, valid-lifetime infinite, preferred-lifetime infinite, prefix fc00:5678::
exchange.go:229: rx 3@10.42.0.23:55505 hello 681µs
```

## check ifconfig
```console
$ kubectl exec vpn-hosta -t -- goes ifconfig
lo[1]: flags=0025<up|loopback|running> type loopback mtu 65536 carrier up group 0
        label lo mode default promiscuity 0 qdisc noqueue qlen 1000 rx-queues 1
        state unknown tx-queues 1 inet 127.0.0.1/8 inet6 ::1/128
eth0[2]: flags=0033<up|broadcast|multicast|running> type ether mtu 1450
        mac 42:99:df:23:65:7b broadcast ff:ff:ff:ff:ff:ff carrier up group 0
        l3broadcast 10.42.0.255 label eth0 link 23 mode default promiscuity 0
        qdisc noqueue qlen 0 rx-queues 2 state up tx-queues 2 inet 10.42.0.23/24
        inet6 fe80::4099:dfff:fe23:657b/64
tun0[3]: flags=0039<up|pointtopoint|multicast|running> type 65534 mtu 1400
        carrier up group 0 mode default peer fc00:1234::10 promiscuity 0
        qdisc fq_codel qlen 500 rx-queues 1 state unknown tx-queues 1
        inet6 fc00:1234::1/128 inet6 fe80::b1d7:552c:c08:2d5/64
```

# ping
```console
$ kubectl exec vpn-hosta -t -- goes ping -c 3 fc00:1234::3
PING fc00:1234::3 (fc00:1234::3); 24 data bytes
32 bytes from fc00:1234::3; icmp_seq=2 ttl=64 time=1.956991ms
```

# delete
```console
$ kubectl delete -f examples/vpn.yaml
```
