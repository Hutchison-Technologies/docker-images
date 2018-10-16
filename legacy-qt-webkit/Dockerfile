FROM hutchisont/legacy-qt:latest

WORKDIR /root
RUN git clone git://code.qt.io/qt/qtwebkit.git && \
    cd qtwebkit/ && \
    qtchooser -qt=qt5 -run-tool=qmake . && \
    make -j8 && \
    cp mkspecs/modules/* /opt/qt/5.7/gcc_64/mkspecs/modules/ && \
    cd /root && \
    git clone https://github.com/zint/zint.git && \
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