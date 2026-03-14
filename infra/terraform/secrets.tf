data "sops_file" "github_app_pem_file" {
  input_type  = "raw"
  source_file = "secrets/restricted/datakit-terraform-github-provider.2024-01-26.private-key.enc.pem"
}

locals {
  github_app_pem_file = data.sops_file.github_app_pem_file.raw
}
