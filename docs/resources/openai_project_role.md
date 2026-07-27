# openai_project_role

API group: [Roles and Assignments](../api-groups/roles-and-assignments.md).

OpenAI API hierarchy: `Administration > Organization > Projects > Roles`.

Manages an OpenAI project custom role. Destroy deletes the custom project role. Use `prevent_destroy` for critical project access roles.

## Example Usage

```terraform
resource "openai_project_role" "api_key_reader" {
  project_id  = openai_project.app.id
  name        = "api-key-reader"
  description = "Read project API keys"
  permissions = [
    "project.api_keys.read",
  ]
}
```

## Import

```sh
terraform import openai_project_role.api_key_reader proj_123/role_123
```

## Schema

### Required

- `project_id` (String) Project ID. Changes replace the role.
- `name` (String) Unique role name.
- `permissions` (Set of String) Permissions granted by the role.

### Optional

- `description` (String) Role description.

### Read-Only

- `id` (String) Role ID.
- `predefined_role` (Boolean) Whether OpenAI predefines the role.
- `resource_type` (String) Role resource type.
