variable "aws_enable_node_monitoring_agent" {
  type        = bool
  description = <<-EOT
  Install the EKS node monitoring agent add-on (`eks-node-monitoring-agent`).

  The agent runs on every worker node and reports additional node health conditions
  (container runtime, kernel, networking, storage). When EKS node auto repair is enabled
  on the node group, it uses them to detect and replace unhealthy nodes beyond the standard
  `Ready` condition.

  The agent runs as a privileged DaemonSet with the `NET_ADMIN` capability, so it is
  disabled by default. Review it against your security requirements before enabling it.

  See https://docs.aws.amazon.com/eks/latest/userguide/node-health.html for more details.

  Example: `true`
  EOT

  default = false
}
