terraform {
  required_providers {
    openai = {
      source = "cloudopsworks/openai"
    }
  }
}

# Initialize from environment variables:
#   OPENAI_ADMIN_KEY or OPENAI_ADMIN_API_KEY - required
#   OPENAI_ORG_ID or OPENAI_ORGANIZATION_ID - optional
#   OPENAI_PROJECT_ID - optional
#   OPENAI_BASE_URL - optional
provider "openai" {}
