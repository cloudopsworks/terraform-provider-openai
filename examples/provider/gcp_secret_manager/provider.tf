terraform {
  required_providers {
    openai = {
      source = "cloudopsworks/openai"
    }
  }
}

provider "openai" {
  gcp_secret_manager = {
    project_id                  = var.gcp_project_id
    secret_id                   = var.openai_secret_id
    version                     = "latest"
    impersonate_service_account = var.openai_secret_reader_service_account
    delegates                   = var.gcp_impersonation_delegates
    scopes                      = ["https://www.googleapis.com/auth/cloud-platform"]
    json_key                    = "admin_api_key"
  }
}

variable "gcp_project_id" {
  type        = string
  description = "GCP project ID that owns the Secret Manager secret."
}

variable "openai_secret_id" {
  type        = string
  description = "GCP Secret Manager secret ID containing OpenAI provider settings."
}

variable "openai_secret_reader_service_account" {
  type        = string
  description = "Service account email to impersonate before reading the secret."
}

variable "gcp_impersonation_delegates" {
  type        = list(string)
  description = "Optional service account impersonation delegate chain."
  default     = []
}
