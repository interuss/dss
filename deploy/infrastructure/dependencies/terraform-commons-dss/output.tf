output "generated_files_location" {
  value = <<-EOT
  Workspace location with generated files: ${local.workspace_location}
  EOT
}

output "workspace_location" {
  value = local.workspace_location
}
