terraform {
  backend "gcs" {
    bucket = "datakit-terraform"
    prefix = "dtkt-integrations/infra"

    kms_encryption_key = "projects/datakit-ops/locations/us/keyRings/secrets/cryptoKeys/terraform"
  }
}

data "terraform_remote_state" "organization" {
  backend = "gcs"

  config = {
    bucket = "datakit-terraform"
    prefix = "datakit/organization"

    kms_encryption_key = "projects/datakit-ops/locations/us/keyRings/secrets/cryptoKeys/terraform"
  }
}

data "terraform_remote_state" "infra" {
  backend = "gcs"

  config = {
    bucket = "datakit-terraform"
    prefix = "datakit/infra"

    kms_encryption_key = "projects/datakit-ops/locations/us/keyRings/secrets/cryptoKeys/terraform"
  }

  workspace = terraform.workspace
}

data "terraform_remote_state" "gcp_github_wif" {
  backend = "gcs"

  config = {
    bucket = "datakit-terraform"
    prefix = "datakit/gcp-github-wif"

    kms_encryption_key = "projects/datakit-ops/locations/us/keyRings/secrets/cryptoKeys/terraform"
  }
}
