FROM google/cloud-sdk:alpine as base

RUN apk add --update ca-certificates git gcc bash libc-dev curl nodejs=20.11.0 build-base npm=10.2.4

RUN echo GIT:$(git version)

RUN echo NPM:$(npm version)

RUN echo NODE:$(node -v)