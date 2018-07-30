# Gcloud + Kubectl + Terraform + Helm

This image is based on the [gcloud-sdk:slim](https://hub.docker.com/r/google/cloud-sdk/) image. On top of this, it has `kubectl`, `terraform`, `helm`, and `helmfile` CLI tools installed.

If you plan to use this image for automation, you may find these posts helpful:

- [Scripting Gcloud](https://cloud.google.com/sdk/docs/scripting-gcloud)
- [Terraform in Automation](https://github.com/mcuadros/terraform-provider-helm/blob/master/vendor/github.com/hashicorp/terraform/website/guides/running-terraform-in-automation.html.md)
- [Helm deployments](https://daemonza.github.io/2017/02/20/using-helm-to-deploy-to-kubernetes/)
