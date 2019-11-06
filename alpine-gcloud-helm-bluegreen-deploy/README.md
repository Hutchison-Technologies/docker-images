# Alpine + Gcloud + Helm + BlueGreen Deploy

This image is based on the [alpine:latest](https://hub.docker.com/_/alpine/) image. On top of this, it has `gcloud`, `kubectl`, and `helm` CLI tools installed.

Includes a "Blue-Green" deployment script which pulls the current "Offline service" to install the helm chart to. 

Once determined, it attempts to apply the update to the offline deployment. 

If successful, it will then redeploy the services with the offline postfix on the service name flipped. It will then scale down the old deployment to zero replica sets. 