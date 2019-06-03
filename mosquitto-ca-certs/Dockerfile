FROM golang:1.11.2-alpine3.8 as cadbuilder

RUN apk add --update \
    coreutils \
    git \
    gcc \
    libc-dev \
    ca-certificates \ 
    wget \
    g++ \
    util-linux-dev \
    make \
    openssl-dev

ENV MOSQUITTO_VERSION v1.4.14
ENV MOSQUITTO_CHECKOUT_DIR /usr/src/mosquitto
RUN mkdir -p ${MOSQUITTO_CHECKOUT_DIR} && \
    git clone https://github.com/eclipse/mosquitto.git ${MOSQUITTO_CHECKOUT_DIR} && \
    cd ${MOSQUITTO_CHECKOUT_DIR} && \
    git checkout ${MOSQUITTO_VERSION} && \
    make clean WITH_SRV=no WITH_UUID=no WITH_DOCS=no && \
    make WITH_SRV=no WITH_UUID=no WITH_DOCS=no && \
    make install WITH_SRV=no WITH_UUID=no WITH_DOCS=no

ENV GLOBAL_RW /rw
RUN mkdir ${GLOBAL_RW}

ENV GLOBAL_SSL_CERT DST_Root_CA_X3.crt
RUN cp /usr/share/ca-certificates/mozilla/${GLOBAL_SSL_CERT} ${GLOBAL_RW}/${GLOBAL_SSL_CERT}

ENV GLOBAL_CA_ROOTS roots.pem
RUN wget -O ${GLOBAL_RW}/${GLOBAL_CA_ROOTS} https://pki.google.com/roots.pem

ENV GLOBAL_SOCK_DIR /var/run/cad
RUN mkdir ${GLOBAL_SOCK_DIR}