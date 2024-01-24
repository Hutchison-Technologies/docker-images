FROM google/cloud-sdk:alpine as base

RUN apk add --update ca-certificates git gcc bash libc-dev curl

COPY --from=golang:1.21.6-alpine3.19 /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

# RUN go install cloud.google.com/go/firestore@v1.14.0

# RUN go install github.com/GoogleCloudPlatform/functions-framework-go@v1.6.1

# RUN go install google.golang.org/api@v0.151.0

# RUN go install github.com/stretchr/testify@v1.8.4