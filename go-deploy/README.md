# Gcloud CLI + Golang v1.25

This image is used to deploy public golang cloud functions to GCP.

Current Prod version tag is `hutchisont/go-deploy:go125_v2.1.1`

## Build and push process (MacOS)

```bash
docker buildx build --platform linux/amd64 -t hutchisont/go-deploy:go125_v2.1.1 . --load
```

```bash
docker tag *imageID* hutchisont/go-deploy:go125_v2.1.1
```

Once tagged, push to dockerhub via:

```bash
docker push hutchisont/go-deploy:go125_v2.1.1
```

## Build and push (Linux)

NOTE: Increment build number for deployments

EG `v1.0.0`

Commands:

```bash
docker build -t hutchisont/go-deploy:go125_v2.1.1 .
```

Once tagged, push to dockerhub via:

```bash
docker push hutchisont/go-deploy:go125_v2.1.1
```
