variable "github_organization_slug" {
  type        = string
  description = "The GitHub organization slug"
}

variable "github_app_id" {
  type        = string
  description = "The GitHub App ID"
}

variable "github_app_installation_id" {
  type        = string
  description = "The GitHub App installation ID"
}

variable "github_organization" {
  description = "The GitHub organization"
  type        = string
}

variable "github_repository" {
  description = "The project GitHub repository"
  type        = string
}

variable "gke_namespace" {
  description = "The project GKE namespace"
  type        = string
}
