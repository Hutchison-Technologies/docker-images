# Gcloud CLI + Golang123

This image is used to deploy public golang cloud functions to GCP.

## Build and push process

```bash
docker build -t hutchisont/go-deploy:tagname .
```

Once tagged, push to dockerhub via:

```bash
docker push hutchisont/go-deploy:tagname
```
