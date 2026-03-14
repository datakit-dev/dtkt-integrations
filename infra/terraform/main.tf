locals {
  project = terraform.workspace == "prd" ? (
    data.terraform_remote_state.organization.outputs.projects.cloud.project
  ) : data.terraform_remote_state.infra.outputs.project

  network       = data.terraform_remote_state.infra.outputs.network
  multi_regions = data.terraform_remote_state.infra.outputs.multi_regions
  locations     = setunion(local.network.regions, local.multi_regions)

  gcp_github_wif_pool_name = data.terraform_remote_state.gcp_github_wif.outputs.pool_name

  gke_project_id = data.terraform_remote_state.infra.outputs.project.id
}
