# openai_admin_api_keys

API group: [Admin API Keys](../api-groups/admin-api-keys.md).

Lists OpenAI organization admin API keys. Values are redacted.

## Example Usage

```terraform
data "openai_admin_api_keys" "all" {
  limit = 100
  order = "desc"
}
```

## Schema

### Optional

- `after` (String) Pagination cursor.
- `limit` (Number) Maximum keys to return.
- `order` (String) Sort order, `asc` or `desc`.

### Read-Only

- `items` (List) Admin API key records.
- `has_more` (Boolean) Whether another page exists.
- `last_id` (String) Last key ID in the page.
