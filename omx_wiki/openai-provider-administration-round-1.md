---
title: "OpenAI Provider Administration Round 1"
tags: ["terraform-provider", "openai", "admin-api", "secrets", "rbac", "documentation", "lessons-learned"]
created: 2026-07-27T13:57:16.248Z
updated: 2026-07-27T17:44:56Z
sources: []
links: ["openai-provider-administration-round-1.md", "openai-provider-progress-checkpoint-2026-07-27.md"]
category: session-log
confidence: medium
schemaVersion: 1
---

# OpenAI Provider Administration Round 1

## Scope captured

This page captures the feature/documentation work completed on branch `feature/provider-implementation-round1` for the OpenAI Terraform/OpenTofu provider. It covers secret-backed provider initialization, OpenAI Administration API resources/data sources, provider usage documentation, validation evidence, and implementation lessons learned.

## Commits produced

- `5875431 feat(provider): support secret-backed initialization`
- `7039984 feat(provider): add admin identity resources`
- `d0a206a docs(provider): document initialization and usage`

Earlier branch commits that are part of the same implementation round:

- `05fd9a2 feat(provider): implement OpenAI administration resources`
- `5e72bab test(provider): cover admin resources and secret resolution`
- `8f8dad9 docs: document OpenAI administration provider usage`

## Work completed

### Secret-backed provider initialization

Implemented provider bootstrap from these sources:

1. Direct provider arguments: `admin_api_key`, `base_url`, `organization_id`, `project_id`.
2. OpenAI environment variables: `OPENAI_ADMIN_KEY`, `OPENAI_ADMIN_API_KEY`, `OPENAI_BASE_URL`, `OPENAI_ORG_ID`, `OPENAI_ORGANIZATION_ID`, `OPENAI_PROJECT_ID`.
3. Exactly one cloud secret source when no direct admin key is available.

Cloud secret sources implemented:

- AWS Secrets Manager with the AWS SDK default credential chain and optional STS AssumeRole (`role_arn`, `role_session_name`, `external_id`, `duration_seconds`).
- GCP Secret Manager with Application Default Credentials and optional service-account impersonation (`impersonate_service_account`, `delegates`, `scopes`).
- Azure Key Vault with `DefaultAzureCredential`.

Secret payload support:

- Plaintext admin API key.
- JSON object with `admin_api_key` or `api_key`, plus optional `base_url`, `organization_id`, and `project_id`.
- Dot-separated `json_key` path for nested JSON when only one key should be extracted.

Provider initialization rules:

- Direct provider configuration and OpenAI environment variables take precedence over secret payload fields.
- If a direct admin key is available, cloud secret sources are ignored for key resolution.
- If no direct admin key is available, multiple configured/discovered cloud secret sources are a deterministic error.
- Raw cloud provider credentials are not Terraform provider arguments; use each cloud SDK's normal credential chain.

### OpenAI Admin API client layer

Expanded `internal/client` with normalized types and OpenAI Go SDK mappings for:

- Organization projects.
- Project service accounts.
- Project API keys owned by service accounts.
- Organization admin API keys.
- Organization users.
- Organization groups.
- Organization group membership.
- Organization roles.
- Project roles.
- User role assignments.
- Group role assignments.

Validation/mapping helpers were added so provider resources operate against normalized internal models rather than raw SDK response structures.

### Terraform provider resources

Resources now registered by the provider:

- `openai_project`
- `openai_service_account`
- `openai_project_api_key`
- `openai_admin_api_key`
- `openai_organization_group`
- `openai_organization_group_user`
- `openai_organization_role`
- `openai_organization_user_role`
- `openai_organization_group_role`
- `openai_project_role`

Important behavior choices:

- `openai_project` destroy archives the project, matching OpenAI API behavior.
- `openai_service_account` creation uses explicit service-account creation without an unmanaged default key; project API keys are issued through `openai_project_api_key`.
- `openai_project_api_key.value` and `openai_admin_api_key.value` are Sensitive and are only available on create. Reads preserve/return metadata and redacted values.
- Admin API keys and critical roles should use `lifecycle { prevent_destroy = true }` when accidental deletion could lock out automation.
- Group membership and role assignment resources use composite import IDs.

Composite import IDs:

- `openai_service_account`: `project_id/service_account_id`
- `openai_project_api_key`: `project_id/service_account_id/api_key_id`
- `openai_organization_group_user`: `group_id/user_id`
- `openai_organization_user_role`: `user_id/role_id`
- `openai_organization_group_role`: `group_id/role_id`
- `openai_project_role`: `project_id/role_id`

### Terraform data sources

Data sources now registered by the provider:

- `openai_project`, `openai_projects`
- `openai_service_account`, `openai_service_accounts`
- `openai_admin_api_key`, `openai_admin_api_keys`
- `openai_organization_user`, `openai_organization_users`
- `openai_organization_group`, `openai_organization_groups`
- `openai_organization_group_user`, `openai_organization_group_users`
- `openai_organization_role`, `openai_organization_roles`
- `openai_organization_user_role`, `openai_organization_user_roles`
- `openai_organization_group_role`, `openai_organization_group_roles`
- `openai_project_role`, `openai_project_roles`

Organization users are intentionally data-source only in this provider version because the OpenAI Go SDK surface exposes user read/list/update/delete but not user creation. User access is managed with group membership and role assignment resources instead of faking an import-only user resource.

### Documentation and samples

README/source docs were expanded with:

- Installation snippet.
- Provider initialization guide.
- Direct/env/AWS/GCP/Azure initialization examples.
- Secret payload formats.
- Provider argument table.
- Resource and data-source inventory.
- Common workflows for project bootstrap, service-account API key creation, organization groups, custom roles, and role assignments.
- Operational guidance and troubleshooting table.

Provider sample files added under `examples/provider/`:

- `examples/provider/provider.tf` — direct variable initialization.
- `examples/provider/environment/provider.tf` — environment-only initialization.
- `examples/provider/aws_secrets_manager/provider.tf` — AWS Secrets Manager with AssumeRole.
- `examples/provider/gcp_secret_manager/provider.tf` — GCP Secret Manager with service-account impersonation.
- `examples/provider/azure_key_vault/provider.tf` — Azure Key Vault.

Resource/data-source docs and examples were added for all new Admin API surfaces. Missing resource docs for the initial project/service-account/API-key resources were also added.

## Validation evidence

Checks run successfully during the implementation/documentation pass:

- `gofmt -w internal/client internal/provider`
- `terraform fmt -recursive examples`
- `terraform fmt -check -recursive examples/provider`
- `ruby -e "require 'yaml'; YAML.load_file('README.yaml')"`
- `make readme`
- `git diff --check`
- `git diff --cached --check`
- `go test ./...`
- `go vet ./...`
- `go build ./...`

`golangci-lint` was checked but not run because it is not installed in the local environment.

## Lessons learned

1. Inspect generated SDK surfaces before designing Terraform resources. The OpenAI Go SDK exposes organization users as read/list/update/delete but not create, so users should remain data sources rather than import-only pseudo-resources.
2. Keep provider bootstrap secrets out of state. The provider resolves secret-manager payloads at configure time and passes only client settings forward; direct/generated OpenAI key resources still require secure Terraform state because key values are stored as Sensitive state attributes.
3. Model cloud identity through native SDK authentication, not provider-specific raw credential fields. AWS assumes roles from the SDK default chain, GCP impersonates from ADC, and Azure uses DefaultAzureCredential. This keeps Terraform configuration portable and avoids duplicating cloud credential handling.
4. Make deletion semantics explicit. OpenAI projects archive on destroy, while keys, groups, memberships, roles, and assignments call delete/remove APIs. Documentation should warn operators and recommend `prevent_destroy` for critical access paths.
5. Composite IDs are essential for nested OpenAI Admin API resources. Import parsing helpers and examples should stay synchronized for project/service-account/key, group/user, user/role, group/role, and project/role resources.
6. Resource docs must explain create-only secrets. Admin and project API key values are available only during create; reads return metadata/redacted values. Operators need to store generated keys in external secret systems after creation.
7. README.md is generated. In this repo, update `README.yaml`, run `make readme`, then validate the generated README. The generator can emit trailing whitespace, so run `git diff --check` and normalize if needed.
8. Keep examples sanitized. Local provider examples, local Terraform CLI config, generated provider binaries, and Terraform state files should not be committed. Stage only generic examples with placeholder IDs and variables.
9. Prefer small normalized provider/client models over exposing SDK response shapes directly. This reduced Terraform schema coupling and made tests easier to fake.
10. Provider/data-source registration counts are useful guardrails. Tests should assert expected resource/data-source counts when adding many surfaces.

## Follow-ups

- Install/run `golangci-lint` locally or in CI for the branch.
- Push `feature/provider-implementation-round1` and open a PR when ready.
- Keep any environment-specific local samples and Terraform state out of git, or move sanitized variants into `examples/` deliberately.
- Revisit organization user resources only if the OpenAI API/SDK later adds user creation semantics.

## Related pages

- [[openai-provider-administration-round-1]]
- [[openai-provider-progress-checkpoint-2026-07-27]]
