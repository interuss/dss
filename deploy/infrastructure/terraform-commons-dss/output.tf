
output "generated_files_location" {
  value = <<-EOT
  Generated files location:
  - workspace: ${local.workspace_location}
  - main.jsonnet: ${abspath(local_file.tanka_config_main.filename)}
  - spec.json: ${abspath(local_file.tanka_config_spec.filename)}
  - make-certs.sh: ${abspath(local_file.make_certs.filename)}
  - apply-certs.sh: ${abspath(local_file.apply_certs.filename)}
  EOT
}