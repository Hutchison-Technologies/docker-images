# go1_27-node20-redocly

This image is built with:

- Node `v20.19.4-alpine`
- Go `v1.27.0-alpine3.23`

## Build and push process

```docker
docker build -t hutchisont/go1_27-node20-redocly .
```

Once tagged, push to dockerhub via:

```docker
docker push hutchisont/go1_27-node20-redocly:latest
```
