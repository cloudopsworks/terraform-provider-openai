# openai_organization_group_users

Lists users in an OpenAI organization group.

## Example Usage

```terraform
data "openai_organization_group_users" "engineering" {
  group_id = "group_123"
}
```

## Schema

### Required

- `group_id` (String) Group ID.

### Optional

- `after`, `limit`, `order` pagination and sort controls.

### Read-Only

- `items` (List) Group users.
- `has_more` (Boolean) Whether another page exists.
- `next` (String) Next cursor.
