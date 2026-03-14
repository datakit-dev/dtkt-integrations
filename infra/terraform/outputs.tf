output "organization" {
  value = data.terraform_remote_state.organization.outputs.organization
}

output "project" {
  value = local.project
}

output "network" {
  value = local.network
}

output "default_multi_region" {
  value = data.terraform_remote_state.infra.outputs.default_multi_region
}

output "gke" {
  value = {
    project_id = local.gke_project_id
    namespace  = var.gke_namespace
  }
}
