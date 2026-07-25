terraform {
  required_providers {
    openai = {
      source = "cloudopsworks/openai"
    }
  }
}

provider "openai" {
  admin_api_key = var.openai_admin_api_key
}

variable "openai_admin_api_key" {
  type      = string
  sensitive = true
}
