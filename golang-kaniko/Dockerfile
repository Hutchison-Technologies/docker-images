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


SHELL ["/bin/bash", "-c"]

RUN docker-credential-gcr configure-docker
RUN echo $PATH
RUN ln $GOPATH/bin/docker-credential-gcr /usr/local/bin
RUN ln $GOPATH/bin/docker-credential-gcr /usr/bin
RUN ls -l $GOPATH/bin | grep docker
RUN ls -l /usr/local/bin | grep docker

RUN docker-credential-gcr configure-docker
