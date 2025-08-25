
## deploy
```console
$ ko apply -L -B --platform=linux/arm64 -f deployments/hello-world.yaml
...
pod/hello-world created
```

## Check logs
```console
$ kubectl logs hello-world
hello world
```

# delete
```console
$ kubectl delete -f deployments/hello-world.yaml
pod "hello-world" deleted
```
