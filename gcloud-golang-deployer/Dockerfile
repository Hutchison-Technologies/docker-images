FROM google/cloud-sdk:alpine as base

RUN apk add --update ca-certificates git gcc bash libc-dev curl

COPY --from=golang:1.16.15-alpine3.15 /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"