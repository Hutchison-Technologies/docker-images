FROM golang:1.16.15-alpine3.15

RUN apk update

RUN apk add yarn

RUN yarn global add serverless

RUN serverless plugin install -n serverless-google-cloudfunctions