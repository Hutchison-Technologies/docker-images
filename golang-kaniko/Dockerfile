FROM gcr.io/kaniko-project/executor:debug-v0.15.0 AS kaniko

FROM scratch
FROM golang:1.18.1

ENV PACKAGES \
    git \
    bash \
    coreutils \
    ca-certificates \  
    curl 

RUN apt install $PACKAGES

RUN go install github.com/GoogleCloudPlatform/docker-credential-gcr@latest 
ENV PATH "$HOME/go/bin:$PATH"

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

# Set the default shell to bash
SHELL ["/bin/bash", "-c"]