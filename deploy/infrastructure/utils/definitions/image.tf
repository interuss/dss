variable "image" {
  type        = string
  description = <<EOT
  URL of the DSS docker image.


  `latest` can be used to use the latest official interuss docker image.
  Official public images are available on Docker Hub: https://hub.docker.com/r/interuss/dss/tags
  See [/build/README.md](../../../../build/README.md#docker-images) Docker images section to learn
  how to build and publish your own image.

  Example: `latest` or `docker.io/interuss/dss:v0.6.0`
  EOT
}