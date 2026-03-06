FROM google/cloud-sdk:alpine AS base

RUN apk add --update ca-certificates git gcc bash libc-dev curl

COPY --from=golang:1.25.7-alpine3.23 /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

# RUN go install cloud.google.com/go/firestore@v1.14.0

# RUN go install github.com/GoogleCloudPlatform/functions-framework-go@v1.6.1

# RUN go install google.golang.org/api@v0.151.0

# RUN go install github.com/stretchr/testify@v1.8.4