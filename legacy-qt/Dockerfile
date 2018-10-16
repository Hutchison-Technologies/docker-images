FROM ubuntu:bionic as installerdeps
RUN apt-get update -q && \
    DEBIAN_FRONTEND=noninteractive apt-get install -q -y --no-install-recommends \
    ca-certificates \
    default-jdk \
    libfontconfig1 \
    libice6 \
    libsm6 \
    libx11-xcb1 \
    libxext6 \
    libxrender1 \
    p7zip \
    xvfb \
    locales \
    build-essential \
    apt-utils \
    sed \
    git \
    cmake \
    ruby \
    gperf \
    bison \
    flex \
    python \
    libgl-dev \
    ruby-dev \
    pkg-config \
    sqlite3 \
    libsqlite3-dev \
    libjpeg-dev \
    zlib1g-dev \
    libpng-dev \
    libicu-dev \
    libxml2-dev \
    libxslt-dev \
    libxcomposite-dev \
    libxrender-dev \
    libwebp-dev \
    libqt5gstreamer-dev \
    libgstreamer1.0-dev \
    libgstreamerd-3-dev \
    libqt5webkit5-dev \
    libglm-dev \
    libboost-all-dev \
    libgstreamermm-1.0-dev \
    libqt5sensors5-dev \
    libqt5webchannel5-dev \
    qtwebengine5-dev \
    libqt5x11extras5-dev \
    libqtgconf-dev \
    qtpositioning5-dev \
    qtbase5-private-dev \
    qtdeclarative5-private-dev \
    libqt5serialport5-dev \
    libqt5opengl5-dev \
    && apt-get clean

ENV QT=5.7.1
ENV QTM=5.7
ENV QTSHA=fdf6b4fb5ee9ade2dec74ddb5bea9e1738911e7ee333b32766c4f6527d185eb4
ENV QTCOMPONENTS=gcc_64
ENV QTRUNFILE=https://download.qt.io/archive/qt/${QTM}/${QT}/qt-opensource-linux-x64-${QT}.run

ADD qt-installer-noninteractive.qs /tmp/qt/script.qs
ADD ${QTRUNFILE} /tmp/qt/installer.run

# use bash for RUN until the next FROM, so we can use bash strict mode
# http://redsymbol.net/articles/unofficial-bash-strict-mode/
SHELL ["/bin/bash", "-c"]
RUN set -euo pipefail \
    && echo "${QTSHA}  /tmp/qt/installer.run" | sha256sum -c \
    && chmod +x /tmp/qt/installer.run \
    && xvfb-run -e /dev/stderr /tmp/qt/installer.run --script /tmp/qt/script.qs \
    && rm -rf /tmp/qt

RUN for COMPONENT in `echo ${QTCOMPONENTS} | tr , ' '`; do echo /opt/qt/${QTM}/${COMPONENT}/lib >> /etc/ld.so.conf.d/qt-${QTM}.conf; done
RUN locale-gen en_US.UTF-8 && DEBIAN_FRONTEND=noninteractive dpkg-reconfigure locales

ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/qt/${QTM}/gcc_64/bin:/opt/qt5/bin
ENV LD_LIBRARY_PATH=/usr/lib/x86_64-linux-gnu/