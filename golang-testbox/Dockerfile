FROM golang:1.21.6-alpine3.19

RUN apk update

RUN apk add nodejs yarn build-base npm

RUN apk add --update ca-certificates git gcc bash libc-dev curl 

RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.56.2
