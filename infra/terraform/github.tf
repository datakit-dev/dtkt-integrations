locals {
  allowed_keys = terraform.workspace == "prd" ? ["confidential", "restricted"] : ["confidential"]
  secret_keys  = { for k, v in data.terraform_remote_state.organization.outputs.secret_keys : k => v if contains(local.allowed_keys, k) }

  push_registry_names = lookup(data.terraform_remote_state.infra.outputs, "docker_registry", [{ push_registry_names = {} }])[0].push_registry_names
  pull_registry_names = lookup(data.terraform_remote_state.infra.outputs, "docker_registry", [{ pull_registry_names = {} }])[0].pull_registry_names
}

resource "google_service_account" "github_actions" {
  project = local.project.id

  account_id   = "dtkt-integrations-gha"
  display_name = "dtkt-integrations-gha"
}

resource "google_service_account_iam_member" "github_actions" {
  service_account_id = google_service_account.github_actions.name

  role   = "roles/iam.workloadIdentityUser"
  member = "principalSet://iam.googleapis.com/${local.gcp_github_wif_pool_name}/attribute.repository_env/${var.github_organization}/${var.github_repository}:${terraform.workspace}"
}

resource "google_kms_crypto_key_iam_member" "cloudkms_cryptoKeyDecrypter" {
  for_each = terraform.workspace == "prd" ? local.secret_keys : { for k, v in local.secret_keys : k => v if k != "restricted" }

  crypto_key_id = each.value.id
  role          = "roles/cloudkms.cryptoKeyDecrypter"
  member        = "serviceAccount:${google_service_account.github_actions.email}"
}

resource "google_project_iam_member" "iam_serviceAccountTokenCreator" {
  project = local.project.id

  role   = "roles/iam.serviceAccountTokenCreator"
  member = "serviceAccount:${google_service_account.github_actions.email}"
}

resource "google_artifact_registry_repository_iam_member" "artifactregistry_createOnPushWriter" {
  for_each = local.push_registry_names

  project    = local.gke_project_id
  location   = each.key
  repository = each.value
  role       = "roles/artifactregistry.createOnPushWriter"
  member     = "serviceAccount:${google_service_account.github_actions.email}"
}

resource "google_artifact_registry_repository_iam_member" "artifactregistry_reader" {
  for_each = local.pull_registry_names

  project    = local.gke_project_id
  location   = each.key
  repository = each.value
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${google_service_account.github_actions.email}"
}

resource "github_actions_variable" "gcp_service_account" {
  repository    = "dtkt-integrations"
  variable_name = "GCP_SERVICE_ACCOUNT_${upper(terraform.workspace)}"
  value         = google_service_account.github_actions.email
}
