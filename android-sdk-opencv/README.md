# Android SDK + Open-CV

This image is used to generate Android APK files with OpenCV dependency.

## Setup

We need to download the following OpenCV zip file for Android builds:

- `https://drive.google.com/file/d/1_LYhWfG-8JgPIxdkgplKetoc3zhQPjBp/view?usp=sharing`

Download this file and drop into the root of the `android-sdk-opencv` folder

## Build and push process

```bash
docker build -t hutchisont/android-sdk-opencv .
```

Once tagged, push to dockerhub via:

```bash
docker push hutchisont/android-sdk-opencv:latest
```
