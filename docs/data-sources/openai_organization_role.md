# openai_organization_role

API group: [Roles and Assignments](../api-groups/roles-and-assignments.md).

OpenAI API hierarchy: `Administration > Organization > Roles`.

Looks up one OpenAI organization role by ID.

## Example Usage

```terraform
data "openai_organization_role" "auditor" {
  id = "role_123"
}
```

## Schema

### Required

- `id` (String) Role ID.

### Read-Only

- `name`, `description`, `permissions`, `predefined_role`, and `resource_type`.
