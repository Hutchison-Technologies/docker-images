FROM golang:1.12.4-alpine3.9

ENV GO111MODULE=off

RUN apk --update --no-cache add ca-certificates wget python curl tar git bash openssh gzip jq gcc build-base
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR ${GOPATH}/src/github.com/Hutchison-Technologies


RUN git clone --depth 1 --no-tags https://github.com/Hutchison-Technologies/helm-deployer.git/
RUN cd helm-deployer && \
    dep ensure -v && \
    go install

FROM alpine:latest
COPY --from=0 /go/bin/helm-deployer /usr/local/bin/
ENV HOME /
ENV PATH /google-cloud-sdk/bin:$PATH
ENV CLOUDSDK_PYTHON_SITEPACKAGES 1
WORKDIR /

RUN apk --update --no-cache add ca-certificates wget python curl tar git bash openssh gzip jq && \
    wget https://dl.google.com/dl/cloudsdk/channels/rapid/google-cloud-sdk.zip && \
    unzip google-cloud-sdk.zip && \
    rm google-cloud-sdk.zip && \
    google-cloud-sdk/install.sh --usage-reporting=false --path-update=true --bash-completion=false --rc-path=/.bashrc --additional-components app kubectl alpha beta && \
    google-cloud-sdk/bin/gcloud config set --installation component_manager/disable_update_check true && \
    curl -fsSL -o /get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 && \
    chmod +x /get_helm.sh && \
    /get_helm.sh --no-sudo && \
    helm init --stable-repo-url https://charts.helm.sh/stable --client-only && \
    mkdir /tmp

COPY gcloud_auth_helm_deploy.sh ./
COPY gcloud_auth_helm_dispel.sh ./
COPY gcloud_auth.sh ./
COPY dispel.sh ./