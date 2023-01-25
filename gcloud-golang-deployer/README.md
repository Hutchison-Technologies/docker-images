# Gcloud + Golang16

This image is used to deploy public golang cloud functions to GCP.

## Build and push process

```bash
➜ docker tag *imageID* hutchisont/gcloud-golang-deployer:tagname
```

Once tagged, push to dockerhub via:

```bash
➜ docker push hutchisont/gcloud-golang-deployer:tagname
```
