## deploy
```console
$ ko apply -L -B --platform=linux/GOARCH -f ./deployments/www-echo.yaml
...
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
$ kubectl -n www-echo exec www-echo -- goes-www-echo ping localhost:8080
hello
```

# delete
```console
$ kubectl delete -f deployments/www-echo.yaml
namespace "www-echo" deleted
pod "www-echo" deleted
```
