# Terraform Provider OpenAI

Terraform/OpenTofu provider for OpenAI Administration APIs.

V1 manages the bootstrap path for one internal application:

- `openai_project`
- `openai_service_account`
- `openai_project_api_key`

The provider consumes an existing OpenAI admin API key. Configure exactly one key source: direct `admin_api_key`, AWS Secrets Manager, GCP Secret Manager, or Azure Key Vault.

See `examples/` for usage. The API key resource exposes the created key material as a Sensitive computed value only at create time; protect Terraform state accordingly.
