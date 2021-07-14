# Alpine + Gcloud + Helm + Helm-Deployer

This image is based on the [alpine:latest](https://hub.docker.com/_/alpine/) image. On top of this, it has `gcloud`, `kubectl`, and `helm` CLI tools installed.

There are additional shell scripts to check the install, log you into gcloud via the SDK.

## Helm Deployer Go Application
This app also pulls in the [hutchison-t/helm-deployer](https://github.com/Hutchison-Technologies/helm-deployer) image from github, and builds it within the docker image, making the gcloud_helm_deploy cli go application available.

If you wish to modify our GreenBlue deploy process, please modify this go service. Once happy with the changes, you'll need to rebuild this image locally.

To build, please build with a --no-cache and --platform linux/amd64 flag like:
```
➜ docker build . --no-cache --platform linux/amd64
```

No-cache is required to ensure that we pull in the latest Go "helm deployer" application.
--platform flag is required to ensure we build for GKE compatible architecture AMD64 rather than ARM, if this is not set then you will see issues in the workflow when delegating agents for jenkins steps, the specific error of the container will be like `standard_init_linux.go:228: exec user process caused: exec format error`

After the build is complete, (this may take a while) tag it like:
```
➜ docker tag *imageID* hutchisont/helm-deployer:latest 
```

Once tagged, push to dockerhub via:
```
➜ docker push hutchisont/helm-deployer 
```

Any jenkins builds which reference this docker image (almost all of them) will now pull in this updated docker image by default.

If you wish to modify Helm Charts and not the deploy process, modify the charts at the repository [hutchison-t/helm-charts](https://github.com/Hutchison-Technologies/helm-charts).