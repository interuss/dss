variable "crdb_enable_liveness_probe" {
  type        = bool
  description = "Set this boolean false to replace the CockroachDB liveness probe by a probe that always succeeds. See https://github.com/interuss/dss/issues/1732 for potential reasons to do so."
  default     = true
}
