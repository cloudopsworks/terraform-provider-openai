terraform {
  required_providers {
    openai = {
      source = "cloudopsworks/openai"
    }
  }
}

provider "openai" {
  aws_secrets_manager = {
    region            = var.aws_region
    secret_id         = var.openai_secret_id
    version_stage     = "AWSCURRENT"
    role_arn          = var.openai_secret_reader_role_arn
    role_session_name = "terraform-provider-openai"
    external_id       = var.aws_external_id
    duration_seconds  = 3600
    json_key          = "admin_api_key"
  }
}

variable "aws_region" {
  type        = string
  description = "AWS region containing the OpenAI admin secret."
}

variable "openai_secret_id" {
  type        = string
  description = "AWS Secrets Manager secret ID or ARN."
}

variable "openai_secret_reader_role_arn" {
  type        = string
  description = "IAM role ARN to assume before reading the secret."
}

variable "aws_external_id" {
  type        = string
  description = "Optional external ID required by the assumed role trust policy."
  default     = null
  sensitive   = true
}
