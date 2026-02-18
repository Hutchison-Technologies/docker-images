FROM  golang:1.25.7-alpine3.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-deploy .

FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:alpine

RUN apk add --update ca-certificates git gcc bash libc-dev curl

COPY --from=golang:1.25.7-alpine3.23 /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b /usr/local/bin

WORKDIR /usr/local/bin

COPY --from=builder /app/go-deploy /usr/local/bin/go-deploy

CMD ["/bin/bash"] 