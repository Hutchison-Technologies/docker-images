ARG jdk=17.0.12_7
ARG android=35

FROM saschpe/android-sdk:${android}-jdk${jdk}

ARG android
ARG cmake=3.22.1
ARG jdk
ARG ndk=27.1.12297006

ENV NDK_ROOT="${ANDROID_SDK_ROOT}/ndk/${ndk}"
RUN sdkmanager --install \
    "cmake;${cmake}" \
    "ndk;${ndk}"

RUN apt-get update && apt-get install -y git nodejs npm yarn unzip