# Build stage 0: install the Qt installer dependencies
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
    && apt-get clean

# Build stage 1: run the Qt installer
FROM installerdeps as qtinstalled

# WARNING: these arguments below MUST be kept up to date by hand for builds
# on Docker Hub to work like they did on your machine, AND MUST match between
# all build stages. Look for them again further below.
ARG QT=5.7.1
ARG QTM=5.7
ARG QTSHA=fdf6b4fb5ee9ade2dec74ddb5bea9e1738911e7ee333b32766c4f6527d185eb4
ARG QTCOMPONENTS=gcc_64

# The rest can be left to the default:
ARG QTRUNFILE=https://download.qt.io/archive/qt/${QTM}/${QT}/qt-opensource-linux-x64-${QT}.run

# Steps, kicking off from the tail end of installerdeps above:
ADD qt-installer-noninteractive.qs /tmp/qt/script.qs
ADD ${QTRUNFILE} /tmp/qt/installer.run
ENV QTM=$QTM
ENV QTSHA=$QTSHA
ENV QTCOMPONENTS=$QTCOMPONENTS
# use bash for RUN until the next FROM, so we can use bash strict mode
# http://redsymbol.net/articles/unofficial-bash-strict-mode/
SHELL ["/bin/bash", "-c"]
RUN set -euo pipefail \
    && echo "${QTSHA}  /tmp/qt/installer.run" | sha256sum -c \
    && chmod +x /tmp/qt/installer.run \
    && xvfb-run -e /dev/stderr /tmp/qt/installer.run --script /tmp/qt/script.qs \
    && rm -rf /tmp/qt

# | egrep -v '\[[0-9]+\] Warning: (Unsupported screen format)|((QPainter|QWidget))' \
# Build stage 2: copy Qt from the first stage, thus needing fewer packages
# and leaving less of a mess e.g. the build layer with /tmp/qt/installer.run
FROM ubuntu:bionic as secondbase

# WARNING: these arguments below MUST be kept up to date by hand for builds
# on Docker Hub to work like they did on your machine, AND MUST match between
# all build stages. Look for them again further above.
ENV QT=5.7.1
ENV QTM=5.7
ENV QTCOMPONENTS=gcc_64

RUN apt-get update -q && \
    DEBIAN_FRONTEND=noninteractive apt-get install -q -y --no-install-recommends \
    locales \
    build-essential \
    p7zip \
    apt-utils \
    && apt-get clean

COPY --from=qtinstalled /opt/qt /opt/qt
RUN for COMPONENT in `echo ${QTCOMPONENTS} | tr , ' '`; do echo /opt/qt/${QTM}/${COMPONENT}/lib >> /etc/ld.so.conf.d/qt-${QTM}.conf; done
RUN locale-gen en_US.UTF-8 && DEBIAN_FRONTEND=noninteractive dpkg-reconfigure locales
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/qt/${QTM}/gcc_64/bin:/opt/qt5/bin
ENV LD_LIBRARY_PATH=/usr/lib/x86_64-linux-gnu/


# FROM secondbase

WORKDIR /root

RUN apt-get update && apt-get install -y \
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
    libqt5serialport5-dev

WORKDIR /root
RUN git clone git://code.qt.io/qt/qtwebkit.git
RUN cd qtwebkit/ && \
    qtchooser -qt=qt5 -run-tool=qmake . && \
    make -j8 && \
    cp mkspecs/modules/* /opt/qt/5.7/gcc_64/mkspecs/modules/

WORKDIR /root
RUN git clone https://github.com/zint/zint.git && \
    cd zint && \
    cmake . && \
    make -j8 && \
    make install

# ENV LEGACY_WORK_DIR /usr/src

# WORKDIR $LEGACY_WORK_DIR

# RUN sed -i "3i #define NO_DB 1" htl_control_application/inc/htl_control_application/ApplicationState.h && \
#     sed -i "3i #define GLM_ENABLE_EXPERIMENTAL 1" tdp_maps/inc/tdp_maps/Map.h && \
#     sed -i "3i #define GLM_ENABLE_EXPERIMENTAL 1" tdp_math_utils/inc/tdp_math_utils/Transformation.h

# RUN for d in */; do cd $d && qtchooser -qt=qt5 -run-tool=qmake . && make clean && make -j8 && make install; done
# RUN cd tdp_qt_widgets && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../tdp_qt_utils && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../htl_zigbee_utils && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../htl_production_db_data && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../htl_common_data_types && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../tdp_crypto_utils && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../tdp_comms && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../boost_system && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../tdp_comms_asio && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../htl_external_api && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../tdp_postgres_db && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../htl_postgres_addon_document && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../htl_production_db && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../htl_production_hardware && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../htl_printing && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../htl_printing_zint && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../htl_production_hardware_printing && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../tdp_math_utils && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../tdp_maps && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../tdp_maps_widget && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../tdp_arduino_utils && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../htl_production_hardware_programming_fixture && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../htl_control_application && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install && \
#     cd ../htl_factory_control_application && \
#     qtchooser -qt=qt5 -run-tool=qmake . && \
#     make clean && \
#     make -j8 && \
#     make install



