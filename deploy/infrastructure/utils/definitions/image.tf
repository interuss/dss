variable "image" {
  type        = string
  description = <<EOT
  URL of the DSS docker image.

  Official public images are available on Docker Hub: https://hub.docker.com/r/interuss/dss/tags
  See [/build/README.md](../../../../build/README.md#docker-images) Docker images section to learn
  how to build and publish your own image.

  Example: `docker.io/interuss/dss:latest` or `docker.io/interuss/dss:v0.14.0`
  EOT

  validation {
    condition = var.image != "latest"
    error_message = "latest value is not supported anymore. Use `docker.io/interuss/dss:latest` for similar behavior."
  }

}
