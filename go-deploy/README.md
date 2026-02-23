# Gcloud CLI + Golang v1.25

This image is used to deploy public golang cloud functions to GCP.

## Build and push process (MacOS)

```bash
docker buildx build --platform linux/amd64 -t hutchisont/go-deploy:tagname . --load
```

```bash
docker tag *imageID* hutchisont/go-deploy:tagname
```

Once tagged, push to dockerhub via:

```bash
docker push hutchisont/go-deploy:tagname
```

## Build and push (Linux)

```bash
docker build -t hutchisont/go-deploy:go125 .
```

Once tagged, push to dockerhub via:

```bash
docker push hutchisont/go-deploy:go125
```
