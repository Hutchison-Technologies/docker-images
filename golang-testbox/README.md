# golang-testbox

This image is built with Golang `v1.21.6`

This image is used as a testbox in our pipeline. It does not have the dependencies installed to deploy code to GCP.

## Build and push process

```docker
docker build -t hutchisont/golang-testbox .
```

Once tagged, push to dockerhub via:

```docker
docker push hutchisont/golang-testbox:latest
```
