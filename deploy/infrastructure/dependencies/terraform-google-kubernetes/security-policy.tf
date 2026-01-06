module "cloud_armor" {
  source  = "GoogleCloudPlatform/cloud-armor/google"
  version = "~> 7.0"

  count = var.external_prometheus == "" ? 0 : 1

  project_id                           = var.google_project_name
  name                                 = "dss-prometheus-security-policy"
  description                          = "DSS: Security policy for external prometheus"
  default_rule_action                  = "deny(403)"
  type                                 = "CLOUD_ARMOR"
  layer_7_ddos_defense_enable          = false
  layer_7_ddos_defense_rule_visibility = "STANDARD"
  json_parsing                         = "STANDARD"
  log_level                            = "NORMAL"

  security_rules = {
    "allowed_subnets" = {
      action        = "allow"
      priority      = 1
      src_ip_ranges = var.external_prometheus_allowed_ips
    }
  }
}
