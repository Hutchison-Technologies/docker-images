FROM golang:1.27.4-alpine3.23

RUN apk update

RUN apk add nodejs yarn build-base npm

RUN apk add --update ca-certificates git gcc bash libc-dev curl 

RUN wget -O- -nv https://golangci-lint.run/install.sh | sh -s v2.15.0
