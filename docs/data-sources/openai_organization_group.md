# openai_organization_group

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
