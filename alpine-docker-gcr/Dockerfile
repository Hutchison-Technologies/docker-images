FROM alpine:3.8

RUN apk --update --no-cache add \
    git \
    openssh \
    bash \
    ca-certificates \
    ncurses-dev \
    openssl-dev \
    pcre \
    unixodbc-dev \
    zlib-dev \
    docker \
    curl \
    python3 \
    tar \
    gzip && \
    pip3 install --upgrade pip && \
    pip3 install docker-compose && \
    curl -L "https://github.com/GoogleCloudPlatform/docker-credential-gcr/releases/download/v1.5.0/docker-credential-gcr_linux_amd64-1.5.0.tar.gz" | \
    tar xz --to-stdout ./docker-credential-gcr > docker-credential-gcr && \
    chmod +x docker-credential-gcr && \
    mv docker-credential-gcr /usr/local/bin