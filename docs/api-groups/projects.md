# Projects API group

Use the Projects API group for OpenAI project lifecycle, project service accounts,
project-scoped service-account API keys, and project custom roles.

## Terraform surfaces

| Surface | Kind | Purpose | Detailed docs |
| --- | --- | --- | --- |
| `openai_project` | Resource | Create and update organization projects. Destroy archives projects. | `docs/resources/openai_project.md` |
| `openai_projects` | Data source | List organization projects, optionally including archived projects. | `docs/data-sources/openai_projects.md` |
| `openai_project` | Data source | Read one project by ID. | `docs/data-sources/openai_project.md` |
| `openai_service_account` | Resource | Create project service accounts without creating an unmanaged default API key. | `docs/resources/openai_service_account.md` |
| `openai_service_accounts` | Data source | List service accounts in a project. | `docs/data-sources/openai_service_accounts.md` |
| `openai_service_account` | Data source | Read one service account by project ID and service account ID. | `docs/data-sources/openai_service_account.md` |
| `openai_project_api_key` | Resource | Issue service-account project API keys with optional create-only scopes. | `docs/resources/openai_project_api_key.md` |
| `openai_project_role` | Resource | Manage project-scoped custom roles. | `docs/resources/openai_project_role.md` |
| `openai_project_roles` | Data source | List project-scoped roles. | `docs/data-sources/openai_project_roles.md` |
| `openai_project_role` | Data source | Read one project-scoped role by ID. | `docs/data-sources/openai_project_role.md` |

## Common workflow

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

## Import and destroy behavior

| Resource | Import ID | Destroy behavior |
| --- | --- | --- |
| `openai_project` | `proj_123` | Archives the project. |
| `openai_service_account` | `proj_123/svc_acct_123` | Deletes the service account. |
| `openai_project_api_key` | `proj_123/svc_acct_123/key_123` | Deletes the key. |
| `openai_project_role` | `proj_123/role_123` | Deletes the project role. |

## Operational notes

- Store `openai_project_api_key.value` only in sensitive outputs or external secret stores.
- Project API key values are returned only at create time and remain in Terraform state.
- Project API key scopes are create-only. Scope changes replace the key.
- Import existing project assets before letting Terraform manage or destroy them.
