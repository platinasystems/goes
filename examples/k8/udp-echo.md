## deploy
```console
$ ko apply -B --platform=linux/arm64 -f ./examples/k8/udp-echo.yaml
2025/08/13 14:57:44 Using base cgr.dev/chainguard/static:latest@sha256:6a4b683f4708f1f167ba218e31fcac0b7515d94c33c3acf223c36d5c6acd3783 for github.com/platinasystems/goes/v2/examples/go-udp-echo
2025/08/13 14:57:45 Building github.com/platinasystems/goes/v2/examples/go-udp-echo for linux/arm64
2025/08/13 14:57:45 Loading ko.local/go-udp-echo:83a9e7661b0b1cda9d30a802230618c3e2a439491e9109815caddf07866b80ba
2025/08/13 14:57:45 Loaded ko.local/go-udp-echo:83a9e7661b0b1cda9d30a802230618c3e2a439491e9109815caddf07866b80ba
2025/08/13 14:57:45 Adding tag latest
2025/08/13 14:57:45 Added tag latest
namespace/udp-echo created
pod/udp-echo created
```

## Check logs
```console
$ kubectl -n udp-echo logs udp-echo
start :7 service
```

# ping
```console
$ kubectl -n udp-echo exec udp-echo -- go-udp-echo ping localhost
0 retransmits, 182077.68778406296pps
```

# delete
```console
$ kubectl delete -f examples/k8/udp-echo.yaml
namespace "udp-echo" deleted
pod "udp-echo" deleted
``
