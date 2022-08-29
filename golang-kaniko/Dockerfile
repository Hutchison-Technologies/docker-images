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

RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.21.4/bin/linux/amd64/kubectl
RUN chmod u+x kubectl && mv kubectl /bin/kubectl

#
# Add kaniko to this image by re-using binaries and steps from official image
#
COPY --from=kaniko /kaniko/ /kaniko/
COPY --from=kaniko /kaniko/warmer /kaniko/warmer
COPY --from=kaniko /kaniko/docker-credential-gcr /kaniko/docker-credential-gcr
COPY --from=kaniko /kaniko/.docker /kaniko/.docker
COPY --from=busybox:1.32.0 /bin /busybox

ENV PATH $PATH:/usr/local/bin:/kaniko:/busybox
ENV DOCKER_CONFIG /kaniko/.docker/
ENV DOCKER_CREDENTIAL_GCR_CONFIG /kaniko/.config/gcloud/docker_credential_gcr_config.json