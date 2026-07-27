terraform {
  required_providers {
    openai = {
      source = "cloudopsworks/openai"
    }
  }
}

provider "openai" {
  azure_key_vault = {
    vault_url   = var.azure_key_vault_url
    secret_name = var.openai_secret_name
    version     = var.openai_secret_version
    json_key    = "admin_api_key"
  }
}

variable "azure_key_vault_url" {
  type        = string
  description = "Azure Key Vault URL, for example https://platform-prod.vault.azure.net/."
}

variable "openai_secret_name" {
  type        = string
  description = "Azure Key Vault secret name containing OpenAI provider settings."
}

variable "openai_secret_version" {
  type        = string
  description = "Optional Azure Key Vault secret version."
  default     = ""
}
