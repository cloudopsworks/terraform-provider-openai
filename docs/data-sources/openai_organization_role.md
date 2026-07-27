# openai_organization_role

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
