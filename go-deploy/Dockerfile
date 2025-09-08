FROM golang:1.23.10-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-deploy .

FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:alpine AS runtime

RUN apk update

RUN apk add --update --no-cache bash ca-certificates git gcc libc-dev curl go

RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b /usr/local/bin

WORKDIR /usr/local/bin

COPY --from=builder /app/go-deploy /usr/local/bin/go-deploy

CMD ["/bin/bash"] 