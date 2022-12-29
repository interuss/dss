variable "image" {
  type        = string
  description = <<EOT
  Full name of the docker image built in the section above. build.sh prints this name as
  the last thing it does when run with DOCKER_URL set. It should look something like
  gcr.io/your-project-id/dss:2020-07-01-46cae72cf if you built the image yourself as
  documented in /build/README.md, or docker.io/interuss/dss.
  EOT
  default     = "docker.io/interuss/dss:v0.6.0"
}