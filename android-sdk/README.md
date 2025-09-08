# Android SDK

This image is used to generate Android APK files.

## Build and push process

```bash
docker build --platform linux/amd64 -t hutchisont/android-sdk .
```

Once tagged, push to dockerhub via:

```bash
docker push hutchisont/android-sdk:latest
```
