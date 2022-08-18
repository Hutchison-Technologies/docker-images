FROM golang:1.16.15-alpine3.15

RUN apk update

RUN apk add yarn

RUN apk add --update ca-certificates git gcc bash libc-dev curl

RUN yarn global add serverless

RUN go get golang.org/x/lint/golint

RUN go get cloud.google.com/go/firestore@v1.6.1

RUN go get github.com/GoogleCloudPlatform/functions-framework-go@v1.6.1

RUN go get google.golang.org/api@v0.90.0

RUN go get github.com/stretchr/testify@v1.8.0