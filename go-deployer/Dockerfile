# Get a Go builder image
FROM golang:1.23.10-alpine3.22 AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary to avoid CGO dependencies in runtime
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-deploy .

# Runtime image with gcloud CLI
FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:alpine AS runtime

# Install bash, go, git, curl, gcc for builds, ca-certificates
RUN apk add --no-cache bash ca-certificates git gcc libc-dev curl go

WORKDIR /workspace

# Install golangci-lint
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s

# Copy the static binary from builder
COPY --from=builder /app/go-deploy /usr/local/bin/go-deploy

# Default to bash shell
CMD ["/bin/bash"]