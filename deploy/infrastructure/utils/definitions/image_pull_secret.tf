variable "image_pull_secret" {
  type        = string
  description = <<-EOT
  Secret name of the credentials to access the image registry.
  If the image specified in `VAR_DOCKER_IMAGE_NAME` requires
  authentication, you can use the following command to store the credentials as secrets:

  > kubectl create secret -n VAR_NAMESPACE docker-registry VAR_DOCKER_IMAGE_PULL_SECRET \
      --docker-server=DOCKER_REGISTRY_SERVER \
      --docker-username=DOCKER_USER \
      --docker-password=DOCKER_PASSWORD \
      --docker-email=DOCKER_EMAIL

  Replace `VAR_DOCKER_IMAGE_PULL_SECRET` with the secret name (for instance: `private-registry-credentials`).
  For docker hub private repository, use `docker.io` as `DOCKER_REGISTRY_SERVER` and an
  [access token](https://hub.docker.com/settings/security) as `DOCKER_PASSWORD`.

  Example: docker-registry
  EOT
  default     = ""
}