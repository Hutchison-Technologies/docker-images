FROM alpine:latest

ENV VERSION v2.15.1
ENV HOME /
ENV PATH /google-cloud-sdk/bin:$PATH
ENV CLOUDSDK_PYTHON_SITEPACKAGES 1
ENV FILENAME helm-${VERSION}-linux-amd64.tar.gz
ENV HELM_URL https://storage.googleapis.com/kubernetes-helm/${FILENAME}

WORKDIR /

RUN apk --update --no-cache add ca-certificates wget python curl tar git bash openssh gzip jq && \
    wget https://dl.google.com/dl/cloudsdk/channels/rapid/google-cloud-sdk.zip && \
    unzip google-cloud-sdk.zip && \
    rm google-cloud-sdk.zip && \
    google-cloud-sdk/install.sh --usage-reporting=false --path-update=true --bash-completion=false --rc-path=/.bashrc --additional-components app kubectl alpha beta && \
    google-cloud-sdk/bin/gcloud config set --installation component_manager/disable_update_check true && \
    curl -o /tmp/$FILENAME ${HELM_URL} && \
    tar -zxvf /tmp/${FILENAME} -C /tmp && \
    mv /tmp/linux-amd64/helm /bin/helm && \
    rm -rf /tmp && \
    helm init --client-only && \
    mkdir /tmp

COPY deploy.sh .
COPY dispel.sh .
COPY gcloud_auth_helm_deploy.sh .