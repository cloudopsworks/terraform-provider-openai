# openai_project_role

API group: [Roles and Assignments](../api-groups/roles-and-assignments.md).

OpenAI API hierarchy: `Administration > Organization > Projects > Roles`.

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
