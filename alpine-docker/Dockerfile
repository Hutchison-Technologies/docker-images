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
    pip3 install docker-compose