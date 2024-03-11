FROM golang:1.21.6-alpine3.19

RUN apk update

RUN apk add nodejs yarn build-base npm

RUN apk add --update ca-certificates git gcc bash libc-dev curl 

RUN yarn global add serverless

RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s

WORKDIR /usr/src/app

RUN go mod init hutchisont/serverlessdeployer

RUN go get cloud.google.com/go/firestore@v1.14.0

RUN go get github.com/GoogleCloudPlatform/functions-framework-go@v1.6.1

RUN go get google.golang.org/api@v0.151.0

RUN go get github.com/stretchr/testify@v1.8.4