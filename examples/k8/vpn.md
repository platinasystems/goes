## deploy
```console
$ ko apply -B --platform=linux/GOARCH -f ./examples/k8/vpn.yaml
2025/08/13 15:44:23 Using base cgr.dev/chainguard/static:latest@sha256:6a4b683f4708f1f167ba218e31fcac0b7515d94c33c3acf223c36d5c6acd3783 for github.com/platinasystems/goes/v2/examples/go-vpn
2025/08/13 15:44:23 Building github.com/platinasystems/goes/v2/examples/go-vpn for linux/arm64
2025/08/13 15:44:24 Loading ko.local/go-vpn:fd86e5c05adbd6d4656c1acebb0de163218d4f9b940aa1337bc590ecc58e4848
2025/08/13 15:44:25 Loaded ko.local/go-vpn:fd86e5c05adbd6d4656c1acebb0de163218d4f9b940aa1337bc590ecc58e4848
2025/08/13 15:44:25 Adding tag latest
2025/08/13 15:44:25 Added tag latest
service/vpn created
secret/registry created
secret/exchange0 created
secret/exchange1 created
secret/exchange2 created
secret/hosta created
secret/hostb created
secret/hostc created
configmap/vpn created
pod/registry created
pod/exchange0 created
pod/hosta created
pod/hostb created
pod/hostc created
```

## Check logs
```console
$ kubectl logs registry
registry/udp.go:30: start receive service
registry/registry.go:214: start main
registry/registry.go:964: start rest :8003
registry/udp.go:40: start send service
registry/registry.go:832: PUT /checkin/exchange/8030
registry/registry.go:401: checked in exchange exchange0,1.1,fc00:1234::1
registry/registry.go:832: PUT /checkin/guest
registry/registry.go:432: checked in guest hostc,6.1,fc00:1234::c
registry/registry.go:832: GET /whois/named/registry
registry/subscriber.go:72: hello from hostc,6.1,fc00:1234::c
registry/registry.go:832: PUT /checkin/guest
registry/registry.go:432: checked in guest hostb,5.1,fc00:1234::b
registry/registry.go:832: GET /whois/named/registry
registry/subscriber.go:72: hello from hostb,5.1,fc00:1234::b
registry/registry.go:832: PUT /checkin/guest
registry/registry.go:432: checked in guest hosta,4.1,fc00:1234::a
registry/registry.go:832: GET /whois/named/registry
registry/subscriber.go:72: hello from hosta,4.1,fc00:1234::a
registry/subscriber.go:72: hello from hostc,6.1,fc00:1234::c,[::ffff:10.42.0.54]:44245
registry/subscriber.go:72: hello from hostb,5.1,fc00:1234::b,[::ffff:10.42.0.55]:34241
```

```console
$ kubectl logs hosta
goresolve.go:62: retry registry.vpn.default.svc.cluster.local
goresolve.go:62: retry registry.vpn.default.svc.cluster.local
goresolve.go:62: retry registry.vpn.default.svc.cluster.local
goresolve.go:62: retry registry.vpn.default.svc.cluster.local
goresolve.go:62: retry registry.vpn.default.svc.cluster.local
guest.go:465: queue whois registry
udp.go:40: start send service
udp.go:30: start receive service
guest.go:162: tun0 up mtu 1412
guest.go:177: tun0 fc00:1234::a%tun0
guest.go:182: tun0 fc00:1234::/64
guest.go:218: start guest 4.1
guest.go:519: start tun0 read service write service
rest.go:664: start hosta whois request service
guest.go:509: start tun0 read service read service
guest.go:405: dropped tun; ip6 8 bytes ff02::2 <- fe80::5e0c:22e1:fc2b:3629; icmp6 sum ok; router-solicitation
subscriber.go:130: resolved registry via 10.42.0.53:8003
subscriber.go:72: hello from registry,0.0,fc00:1234::,10.42.0.53:8003
guest.go:405: dropped tun; ip6 8 bytes ff02::2 <- fe80::5e0c:22e1:fc2b:3629; icmp6 sum ok; router-solicitation
subscriber.go:72: hello from registry,0.0,fc00:1234::,[::ffff:10.42.0.53]:8003
```

# ping
```console
$ kubectl exec hosta -t -- go-vpn ping -c 3 hostb
PING fc00:1234::b (fc00:1234::b); 24 data bytes
32 bytes from fc00:1234::b; icmp_seq=0 ttl=64 time=134.863711ms
32 bytes from fc00:1234::b; icmp_seq=1 ttl=64 time=51.678622ms
32 bytes from fc00:1234::b; icmp_seq=2 ttl=64 time=92.217894ms

--- fc00:1234::b ping statistics ---
3 packets transmitted, 3 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 51.678622ms/92.920076ms/134.863711ms/33.9638ms
```

# delete
```console
$ kubectl delete -f examples/k8/vpn.yaml
```
