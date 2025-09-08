# Gcloud CLI + Golang123

This image is used to deploy public golang cloud functions to GCP.

## Build and push process

```bash
docker buildx build --platform linux/amd64 -t hutchisont/go-deploy:tagname --push .
```
