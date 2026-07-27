# openai_organization_group

API group: [Organization Groups](../api-groups/organization-groups.md).

OpenAI API hierarchy: `Administration > Organization > Groups`.

Looks up one OpenAI organization group by ID.

## Example Usage

```terraform
data "openai_organization_group" "engineering" {
  id = "group_123"
}
```

## Schema

### Required

- `id` (String) Group ID.

### Read-Only

- `name`, `group_type`, `is_scim_managed`, and `created_at`.
