# openai_project_role

Looks up one OpenAI project role by project ID and role ID.

## Example Usage

```terraform
data "openai_project_role" "reader" {
  project_id = "proj_123"
  id         = "role_123"
}
```

## Schema

### Required

- `project_id` (String) Project ID.
- `id` (String) Role ID.

### Read-Only

- `name`, `description`, `permissions`, `predefined_role`, and `resource_type`.
