
## deploy
```console
$ ko apply -B --platform=linux/arm64 -f ./examples/k8/hello-world.yaml
2025/08/13 16:05:21 Using base cgr.dev/chainguard/static:latest@sha256:6a4b683f4708f1f167ba218e31fcac0b7515d94c33c3acf223c36d5c6acd3783 for github.com/platinasystems/goes/v2
2025/08/13 16:05:21 Building github.com/platinasystems/goes/v2 for linux/arm64
2025/08/13 16:05:22 Loading ko.local/goes:043556b7b172a25565e8ef01dd2733defff071a2555d0711ba3f80df1377c508
2025/08/13 16:05:22 Loaded ko.local/goes:043556b7b172a25565e8ef01dd2733defff071a2555d0711ba3f80df1377c508
2025/08/13 16:05:22 Adding tag latest
2025/08/13 16:05:22 Added tag latest
pod/hello-world created
```

## Check logs
```console
$ kubectl logs hello-world
hello world
```

# delete
```console
$ kubectl delete -f examples/k8/hello-world.yaml
pod "hello-world" deleted
```
