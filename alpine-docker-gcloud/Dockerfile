FROM hutchisont/alpine-docker:latest

ENV PATH /google-cloud-sdk/bin:$PATH

WORKDIR /

RUN apk --update --no-cache add ca-certificates wget python curl tar git bash openssh gzip jq perl && \
    wget https://dl.google.com/dl/cloudsdk/channels/rapid/google-cloud-sdk.zip && \
    unzip google-cloud-sdk.zip && \
    rm google-cloud-sdk.zip && \
    google-cloud-sdk/install.sh --usage-reporting=false --path-update=true --bash-completion=false --rc-path=/.bashrc --additional-components app kubectl alpha beta && \
    google-cloud-sdk/bin/gcloud config set --installation component_manager/disable_update_check true

COPY image_name.sh .
COPY enable_private_cluster_access.sh .
COPY restrict_private_cluster_access.sh .