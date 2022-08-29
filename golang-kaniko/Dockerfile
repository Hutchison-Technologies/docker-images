FROM gcr.io/kaniko-project/executor:v1.6.0-debug AS kaniko

FROM scratch
FROM golang:1.16.15-alpine3.15

ENV PACKAGES \
    git \
    bash \
    coreutils \
    ca-certificates \  
    gcc \
    libc-dev \ 
    curl

RUN apk --update-cache add $PACKAGES

COPY --from=kaniko /kaniko/.config /kaniko/.config
COPY --from=kaniko /kaniko/.docker /kaniko/.docker
COPY --from=kaniko /kaniko/executor /kaniko/executor
COPY --from=kaniko /kaniko/ssl /kaniko/ssl
COPY --from=kaniko /kaniko/warmer /kaniko/warmer

ENV HOME /root
ENV USER /root
ENV PATH $PATH:/kaniko
ENV SSL_CERT_DIR /kaniko/ssl/certs
ENV DOCKER_CONFIG /kaniko/.docker/