terraform {
  required_providers {
    openai = {
      source = "cloudopsworks/openai"
    }
  }
}

provider "openai" {
  admin_api_key   = var.openai_admin_api_key
  organization_id = var.openai_organization_id
  project_id      = var.openai_project_id
  base_url        = var.openai_base_url
}

variable "openai_admin_api_key" {
  type        = string
  description = "Existing OpenAI admin API key used to bootstrap provider operations."
  sensitive   = true
}

variable "openai_organization_id" {
  type        = string
  description = "Optional OpenAI organization ID header."
  default     = null
}

variable "openai_project_id" {
  type        = string
  description = "Optional OpenAI project ID header."
  default     = null
}

variable "openai_base_url" {
  type        = string
  description = "Optional OpenAI-compatible base URL override."
  default     = null
}
