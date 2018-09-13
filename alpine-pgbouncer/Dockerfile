FROM alpine:3.8

RUN apk --update add \
    automake \
    build-base \
    libevent-dev \
    curl \
    openssl-dev \
    postgresql-client && \
    curl https://pgbouncer.github.io/downloads/files/1.9.0/pgbouncer-1.9.0.tar.gz | tar xvz && \
    cd pgbouncer-1.9.0 && \
    ./configure --prefix=/usr/local --with-libevent=libevent-prefix && \
    make && \
    make install && \
    apk del git build-base automake autoconf libtool m4 && \
    rm -f /var/cache/apk/* && \
    cd .. && \
    rm -Rf pgbouncer-1.9.0

COPY entrypoint.sh ./

ENTRYPOINT ["./entrypoint.sh"]