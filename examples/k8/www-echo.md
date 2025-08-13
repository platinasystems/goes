## deploy
```console
$ ko apply -B --platform=linux/GOARCH -f ./examples/k8/www-echo.yaml
2025/08/13 16:01:14 Using base cgr.dev/chainguard/static:latest@sha256:6a4b683f4708f1f167ba218e31fcac0b7515d94c33c3acf223c36d5c6acd3783 for github.com/platinasystems/goes/v2/examples/go-www-echo
2025/08/13 16:01:15 Building github.com/platinasystems/goes/v2/examples/go-www-echo for linux/arm64
2025/08/13 16:01:16 Loading ko.local/go-www-echo:36e173f0068b7f02737bde3f34733994cae822b96fc9204b3cea960efe338717
2025/08/13 16:01:16 Loaded ko.local/go-www-echo:36e173f0068b7f02737bde3f34733994cae822b96fc9204b3cea960efe338717
2025/08/13 16:01:16 Adding tag latest
2025/08/13 16:01:16 Added tag latest
namespace/www-echo created
pod/www-echo created
```

## Check logs
```console
$ kubectl logs -n www-echo www-echo
start :8080 service
```

# ping
```console
$ kubectl -n www-echo exec www-echo -- go-www-echo ping localhost:8080
hello
```

# delete
```console
$ kubectl delete -f examples/k8/www-echo.yaml
namespace "www-echo" deleted
pod "www-echo" deleted
```
