# docker-images

Public repo of docker images used in our pipelines.

If you need to update an image, first build it then tag it:

```
➜ docker tag *imageID* hutchisont/[tagName:latest]
```

Once tagged, push to dockerhub via:

```
➜ docker push hutchisont/[tagName]
```
