# Alpine + Docker + GCR + Kubectl

This image is based on the [alpine-docker-gcr](https://hub.docker.com/r/hutchisont/alpine-docker-gcr/) image. On top of this, it has `kubectl` installed.

It also contains an `image_name.sh` script which, given a target env (staging, prod) and an app (example-api), will return the image in use by the production live colour.
