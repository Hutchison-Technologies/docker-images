FROM alpine:3.17.1

RUN apk --update add \
    automake \
    build-base \
    libevent-dev \
    curl \
    openssl-dev \
    postgresql-client \
    gzip \
    tar && \
    curl https://www.pgbouncer.org/downloads/files/1.18.0/pgbouncer-1.18.0.tar.gz | tar -xvz && \
    cd pgbouncer-1.18.0 && \
    ./configure --prefix=/usr/local --with-libevent=libevent-prefix && \
    make && \
    make install && \
    apk del git build-base automake autoconf libtool m4 && \
    rm -f /var/cache/apk/* && \
    cd .. && \
    rm -Rf pgbouncer-1.18.0

COPY entrypoint.sh ./

ENTRYPOINT ["./entrypoint.sh"]