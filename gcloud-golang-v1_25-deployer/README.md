# Gcloud + Golang16

This image is used to deploy public golang cloud functions to GCP.

## Build and push process

```bash
docker build -t hutchisont/gcloud-golang-v1_25-deployer .
```

Once tagged, push to dockerhub via:

```bash
docker push hutchisont/gcloud-golang-v1_25-deployer:latest
```
