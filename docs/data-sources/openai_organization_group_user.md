# openai_organization_group_user

API group: [Organization Groups](../api-groups/organization-groups.md).

OpenAI API hierarchy: `Administration > Organization > Groups > Users`.

Looks up one OpenAI organization group membership by group ID and user ID.

## Example Usage

```terraform
data "openai_organization_group_user" "alice" {
  group_id = "group_123"
  user_id  = "user_123"
}
```

## Schema

### Required

- `group_id` (String) Group ID.
- `user_id` (String) User ID.

### Read-Only

- `id`, `email`, `name`, `is_service_account`, `picture`, and `user_type`.
