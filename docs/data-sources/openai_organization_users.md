# openai_organization_users

API group: [Organization Users](../api-groups/organization-users.md).

Lists OpenAI organization users.

## Example Usage

```terraform
data "openai_organization_users" "alice" {
  emails = ["alice@example.com"]
}
```

## Schema

### Optional

- `after` (String) Pagination cursor.
- `limit` (Number) Maximum users to return.
- `emails` (Set of String) Email filters.

### Read-Only

- `items` (List) Organization users.
- `has_more` (Boolean) Whether another page exists.
- `last_id` (String) Last user ID in the page.
