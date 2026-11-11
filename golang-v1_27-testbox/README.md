# golang-v1_27-testbox

This image is built with Golang `v1.27.0`

NOTE: This should be used for our utils projects where we only run go test and go lint commands;

## Build and push process

```docker
docker build -t hutchisont/golang-v1_27-testbox .
```

Once tagged, push to dockerhub via:

```docker
docker push hutchisont/golang-v1_27-testbox:latest
```
