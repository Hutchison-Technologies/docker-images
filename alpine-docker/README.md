# Alpine + Docker

This image is based on the [alpine:3.8](https://hub.docker.com/_/alpine/) image. On top of this, it has `docker`, `docker-compose`, and `python3` installed.

This image is useful for [building images in CI workflow when you are able to connect to a remote docker host](https://circleci.com/docs/2.0/building-docker-images/).

An example CircleCI `config.yml`:

```
version: 2
jobs:
  build:
    docker:
      - image: hutchisont/alpine-docker:latest
    working_directory: /repo
    steps:
      - checkout
      # ... steps for building/testing app ...

      - setup_remote_docker

      # build and push Docker image
      - run: |
          docker build -t myapp-$CIRCLE_REPO_NAME:$CIRCLE_TAG . && \
          docker push myapp-$CIRCLE_REPO_NAME:$CIRCLE_TAG
```
