FROM google/cloud-sdk:alpine AS base

RUN apk add --update ca-certificates git gcc bash libc-dev curl nodejs build-base npm

RUN echo GIT:$(git version)

RUN echo NPM:$(npm version)

RUN echo NODE:$(node -v)