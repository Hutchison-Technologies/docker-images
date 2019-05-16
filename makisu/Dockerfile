# COPIED from https://github.com/uber/makisu/blob/master/Dockerfile.alpine
FROM golang:1.12 AS builder

RUN mkdir -p /workspace/github.com/uber
WORKDIR /workspace/github.com/uber

RUN git clone --single-branch --branch fix-push-chunk-conf-ignored https://github.com/Rowern/makisu.git
WORKDIR /workspace/github.com/uber/makisu
RUN make vendor && \
    make bins


FROM golang:1.11 AS gcr_cred_helper_builder
RUN go get -u github.com/GoogleCloudPlatform/docker-credential-gcr
RUN CGO_ENABLED=0 make -C /go/src/github.com/GoogleCloudPlatform/docker-credential-gcr && \
    cp /go/src/github.com/GoogleCloudPlatform/docker-credential-gcr/bin/docker-credential-gcr /docker-credential-gcr


FROM golang:1.11 AS ecr_cred_helper_builder
RUN go get -u github.com/awslabs/amazon-ecr-credential-helper/ecr-login/cli/docker-credential-ecr-login
RUN make -C /go/src/github.com/awslabs/amazon-ecr-credential-helper linux-amd64


FROM alpine:3.6
RUN apk add --no-cache libc6-compat

COPY --from=builder /workspace/github.com/uber/makisu/bin/makisu/makisu /makisu-internal/makisu
COPY --from=builder /workspace/github.com/uber/makisu/assets/cacerts.pem /makisu-internal/certs/cacerts.pem

COPY --from=gcr_cred_helper_builder /docker-credential-gcr /makisu-internal/docker-credential-gcr
COPY --from=ecr_cred_helper_builder /go/src/github.com/awslabs/amazon-ecr-credential-helper/bin/linux-amd64/docker-credential-ecr-login /makisu-internal/docker-credential-ecr-login
ENTRYPOINT ["/makisu-internal/makisu"]