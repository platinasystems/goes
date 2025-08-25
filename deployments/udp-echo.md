## deploy
```console
$ ko apply -L -B --platform=linux/arm64 -f deployments/udp-echo.yaml
...
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
$ kubectl -n udp-echo exec udp-echo -- goes-udp-echo ping localhost
0 retransmits, 182077.68778406296pps
```

# delete
```console
$ kubectl delete -f deployments/udp-echo.yaml
namespace "udp-echo" deleted
pod "udp-echo" deleted
``
