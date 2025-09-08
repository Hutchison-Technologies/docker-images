# Build the Go binary
FROM  golang:1.23.10-alpine3.22 AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN go build -o go-deploy .

# Runtime with gcloud
FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:alpine AS base

RUN apk add --update ca-certificates git gcc bash libc-dev curl

COPY --from=golang:1.23.10-alpine3.22 /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /workspace

# Copy the binary
COPY --from=builder /app/go-deploy /usr/local/bin/go-deploy

ENTRYPOINT ["go-deploy"]