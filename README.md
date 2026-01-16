# docker-images

Public repo of docker images used in our pipelines.

If you need to update an image, first build and tag it:

```cmd
docker build . -t hutchisont/[tagName:latest]
```

EG:

```cmd
docker build . -t hutchisont/gcloud-npm
```

Once tagged, push to dockerhub via:

```cmd
docker push hutchisont/[tagName]
```

EG:

```cmd
docker push hutchisont/gcloud-npm
```
