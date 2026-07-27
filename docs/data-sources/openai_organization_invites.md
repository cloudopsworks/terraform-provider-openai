# openai_organization_invites

API group: [Organization Controls](../api-groups/organization-controls.md).

Lists OpenAI organization invites.

## Example Usage

```terraform
data "openai_organization_invites" "pending" {
  limit = 100
}
```

## Schema

### Optional

- `after` (String) Pagination cursor.
- `limit` (Number) Maximum invites to return.

### Read-Only

- `items` (List) Invite records.
- `has_more` (Boolean) Whether another page exists.
- `last_id` (String) Last invite ID in the page.
