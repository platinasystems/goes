
## deploy
```console
$ ko apply -L -B --platform=linux/arm64 -f deployments/standby.yaml
...
pod/standby created
```

## Check logs
```console
$ kubectl logs standby
```

## exec
```console
$ kubectl exec standby -t -- goes echo hello world
hello world
```

# delete
```console
$ kubectl delete -f deployments/standby.yaml
pod "standby" deleted
```
