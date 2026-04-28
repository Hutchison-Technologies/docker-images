FROM node:22.22.1-alpine AS nodebase

FROM google/cloud-sdk:alpine AS base

RUN apk add --update ca-certificates git gcc bash libc-dev curl build-base

COPY --from=nodebase /usr/local /usr/local

RUN echo NODE:$(node -v)

RUN echo NPM:$(npm version)