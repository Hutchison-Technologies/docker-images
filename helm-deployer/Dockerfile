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

RUN apk --update --no-cache add ca-certificates wget py-pip curl tar git bash openssl openssh gzip jq && \   
    wget https://dl.google.com/dl/cloudsdk/channels/rapid/google-cloud-sdk.zip && \
    unzip google-cloud-sdk.zip && \
    rm google-cloud-sdk.zip && \
    google-cloud-sdk/install.sh --usage-reporting=false --path-update=true --bash-completion=false --rc-path=/.bashrc --additional-components alpha beta && \
    google-cloud-sdk/bin/gcloud config set --installation component_manager/disable_update_check true && \
    https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash && \
    curl -LO 'https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl' && \
    install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

COPY gcloud_auth_helm_deploy.sh ./
COPY gcloud_auth_helm_dispel.sh ./
COPY gcloud_auth.sh ./
COPY dispel.sh ./