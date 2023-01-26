FROM google/cloud-sdk:alpine as base

RUN apk add --update ca-certificates git gcc bash libc-dev curl

COPY --from=golang:1.16.15-alpine3.15 /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

RUN go get cloud.google.com/go/firestore@v1.9.0

RUN go get github.com/GoogleCloudPlatform/functions-framework-go@v1.6.1

RUN go get google.golang.org/api@v0.90.0

RUN go get github.com/stretchr/testify@v1.8.1