# openai_project

API group: [Projects](../api-groups/projects.md).

Manages an OpenAI organization project through the Administration API. Destroy
archives the project because OpenAI projects are archived rather than hard-deleted.

Use projects to isolate applications, teams, environments, or billing boundaries
inside an OpenAI organization. Service accounts, project API keys, and project
roles are scoped to a project.

## Example Usage

```terraform
resource "openai_project" "app" {
  name = "example-app"
}
```

Optional external key and geography:

```terraform
resource "openai_project" "app" {
  name            = "search-prod"
  external_key_id = "cost-center-1234"
  geography       = "us"
}
```

## Import

```sh
terraform import openai_project.app proj_123
```

## Schema

### Required

- `name` (String) Project display name.

### Optional

- `external_key_id` (String) External key ID associated with the project.
- `geography` (String) Optional data residency geography for organizations where OpenAI enables geography selection. Changes replace the project.

### Read-Only

- `id` (String) OpenAI project ID.
- `status` (String) Project status, such as `active` or `archived`.
- `created_at` (Number) Unix timestamp for project creation.
- `archived_at` (Number) Unix timestamp for project archival when archived.

## Notes

- Terraform destroy archives the project in OpenAI.
- Import existing projects before managing them to avoid duplicate project names.
- Use `prevent_destroy` on critical production projects if accidental archival would cause downtime.
