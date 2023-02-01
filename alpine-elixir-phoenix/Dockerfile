FROM elixir:1.7.2-alpine

WORKDIR /usr/app

ENV PHX_NEW_VER 1.3.4
ENV PHX_NEW_HASH "70ceb577750c1b4dc417cd6ec1c02fea3b31c6ba55aa7fd366253ca4acce81f7b1127a5917b03f32ff7ff3516afa985bc5c5582ce0de8fa3668653b83be91903"

RUN apk add --update --no-cache inotify-tools \
    git \
    openssh \
    bash \
    ca-certificates \
    ncurses-dev \
    openssl-dev \
    pcre \
    unixodbc-dev \
    zlib-dev \
    tar \
    gzip \
    make \
    gcc \
    erlang-dev \
    build-base && \
    mix local.hex --force && \
    mix local.rebar --force && \
    mix archive.install https://github.com/phoenixframework/archives/raw/master/phx_new-${PHX_NEW_VER}.ez --force --sha512 ${PHX_NEW_HASH}
COPY wait-for /