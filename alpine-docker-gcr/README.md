# Alpine + Docker

This image is based on the [alpine:3.8](https://hub.docker.com/_/alpine/) image. On top of this, it has `docker`, `docker-compose`, `python3`, and `docker-credential-gcr` installed.

This image is useful for [building and pushing images to GCR in CI workflow when you are able to connect to a remote docker host](https://circleci.com/docs/2.0/building-docker-images/).
