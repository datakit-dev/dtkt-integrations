terraform {
  required_providers {
    sops = {
      source  = "carlpett/sops"
      version = "1.4.0"
    }

    github = {
      source  = "integrations/github"
      version = "6.6.0"
    }
  }
}

provider "sops" {}

# https://registry.terraform.io/providers/integrations/github/latest/docs
provider "github" {
  owner = var.github_organization_slug
  app_auth {
    id              = var.github_app_id
    installation_id = var.github_app_installation_id
    pem_file        = local.github_app_pem_file
  }
}
