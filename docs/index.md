# OpenAI provider

## Installation

Declare the provider in the root module. Pin a version in production once a release is selected.

```hcl
terraform {
  required_providers {
    openai = {
      source = "cloudopsworks/openai"
      # version = "~> 0.1"
    }
  }
}
```

## Provider initialization

The provider configures an OpenAI Administration API client. Every resource and
data source requires an OpenAI admin API key supplied directly, by environment
variable, or by resolving exactly one supported secret manager.

Resolution order:

1. Explicit provider arguments: `admin_api_key`, `base_url`, `organization_id`, and `project_id`.
2. OpenAI environment variables: `OPENAI_ADMIN_KEY` or `OPENAI_ADMIN_API_KEY`, `OPENAI_BASE_URL`, `OPENAI_ORG_ID` or `OPENAI_ORGANIZATION_ID`, and `OPENAI_PROJECT_ID`.
3. One cloud secret source when no direct admin key is configured.

Direct provider arguments and OpenAI environment variables override fields read
from a secret payload. If `admin_api_key` or `OPENAI_ADMIN_KEY` /
`OPENAI_ADMIN_API_KEY` is set, cloud secret-manager blocks are ignored for key
resolution.

### Direct initialization

Use this for local development, secure CI variables, or platforms that inject the
admin key as a sensitive Terraform variable.

```hcl
provider "openai" {
  admin_api_key   = var.openai_admin_api_key
  organization_id = var.openai_organization_id # optional
  project_id      = var.openai_project_id      # optional
  base_url        = var.openai_base_url        # optional, for compatible gateways/tests
}

variable "openai_admin_api_key" {
  type      = string
  sensitive = true
}
```

### Environment-only initialization

When the provider block is empty, the provider reads OpenAI environment
variables. This is useful in CI runners and local shells where secrets are
already managed outside Terraform.

```hcl
provider "openai" {}
```

```sh
export OPENAI_ADMIN_KEY="sk-admin-..."
export OPENAI_ORG_ID="org_..."        # optional
export OPENAI_PROJECT_ID="proj_..."   # optional
export OPENAI_BASE_URL="https://api.openai.com/v1" # optional
```

### AWS Secrets Manager initialization with AssumeRole

The AWS source uses the AWS SDK default credential chain as the base identity,
then optionally assumes `role_arn` before reading the secret. Do not pass raw AWS
access keys to this provider; use the standard AWS environment, profile, web
identity, instance metadata, or other SDK-supported authentication sources.

```hcl
provider "openai" {
  aws_secrets_manager = {
    region            = "us-east-1"
    secret_id         = "prod/openai/admin"
    version_stage     = "AWSCURRENT"
    role_arn          = "arn:aws:iam::123456789012:role/openai-secret-reader"
    role_session_name = "terraform-provider-openai"
    external_id       = var.aws_external_id
    duration_seconds  = 3600
    json_key          = "admin_api_key"
  }
}
```

AWS secret-source environment discovery is also supported when no direct admin
key and no HCL secret-source block is configured:

| Purpose | Environment variables |
| --- | --- |
| Secret ID | `OPENAI_AWS_SECRETS_MANAGER_SECRET_ID`, `OPENAI_AWS_SECRET_ID` |
| Region | `OPENAI_AWS_SECRETS_MANAGER_REGION`, `AWS_REGION`, `AWS_DEFAULT_REGION` |
| Version | `OPENAI_AWS_SECRETS_MANAGER_VERSION_ID`, `OPENAI_AWS_SECRETS_MANAGER_VERSION_STAGE` |
| JSON key | `OPENAI_AWS_SECRETS_MANAGER_JSON_KEY` |
| Assume role | `OPENAI_AWS_SECRETS_MANAGER_ROLE_ARN`, `OPENAI_AWS_ROLE_ARN` |
| Session name | `OPENAI_AWS_SECRETS_MANAGER_ROLE_SESSION_NAME`, `OPENAI_AWS_ROLE_SESSION_NAME` |
| External ID | `OPENAI_AWS_SECRETS_MANAGER_EXTERNAL_ID`, `OPENAI_AWS_EXTERNAL_ID` |
| Session duration | `OPENAI_AWS_SECRETS_MANAGER_DURATION_SECONDS`, `OPENAI_AWS_ROLE_DURATION_SECONDS` |

### GCP Secret Manager initialization with service-account impersonation

The GCP source uses Application Default Credentials as the base identity and can
impersonate a target service account before reading the secret.

```hcl
provider "openai" {
  gcp_secret_manager = {
    project_id                  = "platform-prod"
    secret_id                   = "openai-admin-key"
    version                     = "latest"
    impersonate_service_account = "openai-secret-reader@platform-prod.iam.gserviceaccount.com"
    delegates                   = []
    scopes                      = ["https://www.googleapis.com/auth/cloud-platform"]
    json_key                    = "admin_api_key"
  }
}
```

GCP secret-source environment discovery is supported when no direct admin key and
no HCL secret-source block is configured:

| Purpose | Environment variables |
| --- | --- |
| Project ID | `OPENAI_GCP_SECRET_MANAGER_PROJECT_ID`, `GOOGLE_CLOUD_PROJECT`, `GCLOUD_PROJECT` |
| Secret ID | `OPENAI_GCP_SECRET_MANAGER_SECRET_ID`, `OPENAI_GCP_SECRET_ID` |
| Version | `OPENAI_GCP_SECRET_MANAGER_VERSION` |
| JSON key | `OPENAI_GCP_SECRET_MANAGER_JSON_KEY` |
| Impersonated service account | `OPENAI_GCP_SECRET_MANAGER_IMPERSONATE_SERVICE_ACCOUNT`, `OPENAI_GCP_IMPERSONATE_SERVICE_ACCOUNT` |
| Delegates | `OPENAI_GCP_SECRET_MANAGER_DELEGATES` (comma-separated) |
| OAuth scopes | `OPENAI_GCP_SECRET_MANAGER_SCOPES` (comma-separated) |

### Azure Key Vault initialization

The Azure source reads a Key Vault secret using `DefaultAzureCredential`.
Configure Azure authentication using the standard Azure SDK mechanisms for your
environment, then point the provider at the vault and secret name.

```hcl
provider "openai" {
  azure_key_vault = {
    vault_url   = "https://platform-prod.vault.azure.net/"
    secret_name = "openai-admin-key"
    version     = "" # optional; omit or leave empty for latest
    json_key    = "admin_api_key"
  }
}
```

### Secret payload formats

Secret values can be plaintext or JSON.

Plaintext secret:

```text
sk-admin-...
```

JSON secret without `json_key`:

```json
{
  "admin_api_key": "sk-admin-...",
  "base_url": "https://api.openai.com/v1",
  "organization_id": "org_...",
  "project_id": "proj_..."
}
```

JSON secret with `json_key = "openai.admin_api_key"`:

```json
{
  "openai": {
    "admin_api_key": "sk-admin-..."
  }
}
```

When `json_key` is set, only that key path is used as the admin API key. When it
is omitted and the payload is a JSON object, the provider reads
`admin_api_key`/`api_key` and optional `base_url`, `organization_id`, and
`project_id` aliases from the object.

## Provider arguments

| Argument | Required | Sensitive | Description |
| --- | --- | --- | --- |
| `admin_api_key` | No | Yes | Existing OpenAI admin API key. Highest-precedence key source. |
| `base_url` | No | No | Optional OpenAI API base URL override for tests or compatible/private gateways. |
| `organization_id` | No | No | Optional organization header value. |
| `project_id` | No | No | Optional project header value. |
| `aws_secrets_manager` | No | No | Reads provider settings from AWS Secrets Manager; supports `role_arn` assume-role. |
| `gcp_secret_manager` | No | No | Reads provider settings from GCP Secret Manager; supports service-account impersonation. |
| `azure_key_vault` | No | No | Reads provider settings from Azure Key Vault. |

Configure only one cloud secret source when no direct admin key is present. The
provider fails fast if multiple secret sources are configured or discovered.

## API group documentation

The provider documentation is organized by OpenAI Administration API group. Use
the catalog below to choose the right Terraform surface, then open the detailed
resource/data-source page or the group guide under `docs/api-groups/`.

### Projects API group

Use this group to manage project containers, service accounts, project-scoped
service-account API keys, and project custom roles.

| Terraform surface | Kind | Purpose | Import ID / Lookup |
| --- | --- | --- | --- |
| `openai_project` | Resource | Create and update organization projects. Destroy archives the project. | `proj_123` |
| `openai_projects` / `openai_project` | Data sources | List projects or read a project by ID. | `id` |
| `openai_service_account` | Resource | Create project service accounts without implicit default keys. | `proj_123/svc_acct_123` |
| `openai_service_accounts` / `openai_service_account` | Data sources | List or read service accounts for a project. | `project_id`, `id` |
| `openai_project_api_key` | Resource | Issue service-account project API keys with optional create-only scopes. | `proj_123/svc_acct_123/key_123` |
| `openai_project_role` | Resource | Manage project-scoped custom roles. | `proj_123/role_123` |
| `openai_project_roles` / `openai_project_role` | Data sources | List or read project-scoped roles. | `project_id`, `id` |

Guide: `docs/api-groups/projects.md`.

### Admin API Keys API group

Use this group to rotate or manage organization admin API keys used for
administration automation.

| Terraform surface | Kind | Purpose | Import ID / Lookup |
| --- | --- | --- | --- |
| `openai_admin_api_key` | Resource | Issue organization admin API keys with optional expiration and scopes. Destroy revokes the key. | `admin_key_123` |
| `openai_admin_api_keys` / `openai_admin_api_key` | Data sources | List or read admin key metadata. Values are redacted. | `id` |

`openai_admin_api_key.value` is Sensitive and returned only during create. It is
still stored in Terraform state; use a remote encrypted backend with restricted
access. Configure at most one of `expires_in_seconds`, `expire_in_hours`, or
`expire_in_days`; hours and days are converted to OpenAI's seconds-based create
parameter. Optional `scopes` are create-only strings sent as the create request
`scopes` array; changing scopes replaces the key. Imported keys should leave
`scopes` unset unless replacement is intended.

Destroy and replacement require OpenAI to confirm `deleted=true` for the
requested admin key ID. If OpenAI does not confirm deletion, Terraform keeps the
resource in state and reports a revoke failure.

Guide: `docs/api-groups/admin-api-keys.md`.

### Organization Users API group

Use this group to discover existing organization users before assigning groups
or roles. Users are data-source only in this provider version; invite and user
lifecycle management stays outside Terraform here.

| Terraform surface | Kind | Purpose | Lookup |
| --- | --- | --- | --- |
| `openai_organization_user` | Data source | Read a user by ID, including profile and project summary metadata. | `id` |
| `openai_organization_users` | Data source | List users, optionally filtered by email addresses. | `emails`, `limit`, `after` |

Guide: `docs/api-groups/organization-users.md`.

### Organization Groups API group

Use this group to manage organization groups and explicit group membership.

| Terraform surface | Kind | Purpose | Import ID / Lookup |
| --- | --- | --- | --- |
| `openai_organization_group` | Resource | Create and update organization groups. | `group_123` |
| `openai_organization_groups` / `openai_organization_group` | Data sources | List or read organization groups. | `id` |
| `openai_organization_group_user` | Resource | Add a user to a group. Destroy removes membership only. | `group_123/user_123` |
| `openai_organization_group_users` / `openai_organization_group_user` | Data sources | List or read group members. | `group_id`, `user_id` |

Guide: `docs/api-groups/organization-groups.md`.

### Roles and Assignments API group

Use this group to define organization/project custom roles and assign
organization roles to users or groups.

| Terraform surface | Kind | Purpose | Import ID / Lookup |
| --- | --- | --- | --- |
| `openai_organization_role` | Resource | Manage organization-scoped custom roles. | `role_123` |
| `openai_organization_roles` / `openai_organization_role` | Data sources | List or read organization roles. | `id` |
| `openai_project_role` | Resource | Manage project-scoped custom roles. | `proj_123/role_123` |
| `openai_project_roles` / `openai_project_role` | Data sources | List or read project roles. | `project_id`, `id` |
| `openai_organization_user_role` | Resource | Assign an organization role to a user. Destroy removes the assignment. | `user_123/role_123` |
| `openai_organization_user_roles` / `openai_organization_user_role` | Data sources | List or read user role assignments. | `user_id`, `role_id` |
| `openai_organization_group_role` | Resource | Assign an organization role to a group. Destroy removes the assignment. | `group_123/role_123` |
| `openai_organization_group_roles` / `openai_organization_group_role` | Data sources | List or read group role assignments. | `group_id`, `role_id` |

Guide: `docs/api-groups/roles-and-assignments.md`.

## Common workflows by API group

### Projects: bootstrap a project and issue a service-account API key

```hcl
resource "openai_project" "app" {
  name            = "search-prod"
  external_key_id = "cost-center-1234"
}

resource "openai_service_account" "app" {
  project_id = openai_project.app.id
  name       = "search-prod"
  role       = "member"
}

resource "openai_project_api_key" "app" {
  project_id         = openai_project.app.id
  service_account_id = openai_service_account.app.id
  name               = "search-prod"
  scopes             = ["responses.read"]
}

output "project_api_key" {
  value     = openai_project_api_key.app.value
  sensitive = true
}
```

### Admin API Keys: create scoped automation credentials

```hcl
resource "openai_admin_api_key" "readonly_admin" {
  name           = "readonly-admin-automation"
  expire_in_days = 30
  scopes = [
    "organization.users.read",
    "organization.projects.read",
  ]

  lifecycle {
    prevent_destroy = true
  }
}
```

Rotate automation credentials by creating a new `openai_admin_api_key` with a
distinct name, storing the new `value` in your external secret manager or CI
secret store, updating provider initialization to read the new key, running
`terraform plan`, and then deliberately removing or destroying the old key. For
critical keys, add `prevent_destroy` and remove that lifecycle protection only
as part of an approved rotation change.

### Organization Users and Groups: resolve users and manage membership

```hcl
data "openai_organization_user" "alice" {
  id = "user_123"
}

resource "openai_organization_group" "engineering" {
  name = "Engineering"
}

resource "openai_organization_group_user" "engineering_alice" {
  group_id = openai_organization_group.engineering.id
  user_id  = data.openai_organization_user.alice.id
}
```

### Roles and Assignments: create roles and bind them to principals

```hcl
resource "openai_organization_role" "auditor" {
  name        = "auditor"
  description = "Read-only organization auditor"
  permissions = [
    "organization.users.read",
  ]
}

resource "openai_organization_group_role" "engineering_auditor" {
  group_id = openai_organization_group.engineering.id
  role_id  = openai_organization_role.auditor.id
}

resource "openai_project_role" "api_key_reader" {
  project_id  = openai_project.app.id
  name        = "api-key-reader"
  description = "Read project API keys"
  permissions = [
    "project.api_keys.read",
  ]
}
```

## Operational guidance

- Prefer a short-lived bootstrap admin key stored in AWS Secrets Manager, GCP Secret Manager, or Azure Key Vault instead of committing direct credentials to Terraform variable files.
- Keep generated OpenAI key values out of logs. Terraform marks them Sensitive, but state backends and plan artifacts still require access control.
- Use remote encrypted state with state locking for shared administration.
- Use provider aliases when managing multiple OpenAI organizations or bootstrap identities from the same root module.
- Import existing projects, service accounts, roles, groups, and keys before letting Terraform manage them.
- Keep organization users managed outside this provider, then reference them with `openai_organization_user` or `openai_organization_users` data sources.
- Review destructive plans carefully: project destroy archives projects, while key, group, membership, role, and role-assignment destroys call the corresponding OpenAI delete/remove API.
- Admin API key destroy/replacement requires OpenAI to confirm `deleted=true` for the requested key ID; otherwise Terraform keeps the resource in state and reports a revoke failure.

Provider alias example:

```hcl
provider "openai" {
  alias         = "prod"
  admin_api_key = var.prod_openai_admin_api_key
}

provider "openai" {
  alias         = "sandbox"
  admin_api_key = var.sandbox_openai_admin_api_key
}

resource "openai_project" "prod_app" {
  provider = openai.prod
  name     = "search-prod"
}
```

## Troubleshooting

| Symptom | What to check |
| --- | --- |
| `missing OpenAI admin API key source` | Set `admin_api_key`, `OPENAI_ADMIN_KEY`/`OPENAI_ADMIN_API_KEY`, or exactly one secret-source block. |
| `ambiguous OpenAI admin API key source` | Remove extra `aws_secrets_manager`, `gcp_secret_manager`, or `azure_key_vault` blocks/env discovery. |
| AWS assume-role validation error | `role_session_name`, `external_id`, and `duration_seconds` require `role_arn`. Duration must be positive. |
| GCP delegates validation error | `delegates` require `impersonate_service_account`. |
| JSON secret error | Verify the payload is valid JSON, `json_key` has no empty path segment, and the selected key resolves to a non-empty string. |
| Data source not found | Confirm the ID belongs to the expected organization/project and that the bootstrap admin key has permission to read it. |
