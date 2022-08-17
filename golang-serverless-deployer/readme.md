# Golang16 + Serverless

This image is used to deploy serverless cloud functions to GCP. It contains an install of golang16 and the [serverless-framework](https://www.serverless.com/).

## Build and push process

```
➜ docker tag *imageID* hutchisont/golang-serverless-deployer:tagname
```

Once tagged, push to dockerhub via:

```
➜ docker push hutchisont/golang-serverless-deployer:tagname
```
