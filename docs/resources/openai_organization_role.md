# openai_organization_role

API group: [Roles and Assignments](../api-groups/roles-and-assignments.md).

OpenAI API hierarchy: `Administration > Organization > Roles`.

Manages an OpenAI organization custom role. Destroy deletes the custom role. Use `prevent_destroy` for critical access roles.

## Example Usage

```terraform
resource "openai_organization_role" "auditor" {
  name        = "auditor"
  description = "Read-only organization auditor"
  permissions = [
    "organization.users.read",
  ]
}
```

## Import

```sh
terraform import openai_organization_role.auditor role_123
```

## Schema

### Required

- `name` (String) Unique role name.
- `permissions` (Set of String) Permissions granted by the role.

### Optional

- `description` (String) Role description.

### Read-Only

- `id` (String) Role ID.
- `predefined_role` (Boolean) Whether OpenAI predefines the role.
- `resource_type` (String) Role resource type.
