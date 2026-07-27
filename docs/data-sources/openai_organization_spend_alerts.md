# openai_organization_spend_alerts

API group: [Organization Controls](../api-groups/organization-controls.md).

OpenAI API hierarchy: `Administration > Organization > Spend Alerts`.

Lists OpenAI organization spend alerts.

## Example Usage

```terraform
data "openai_organization_spend_alerts" "all" {
  limit = 100
  order = "desc"
}
```

## Schema

### Optional

- `after` (String) Cursor for the next page.
- `before` (String) Cursor for the previous page.
- `limit` (Number) Maximum spend alerts to return.
- `order` (String) Sort order by creation time: `asc` or `desc`.

### Read-Only

- `items` (List) Spend alert records.
- `has_more` (Boolean) Whether another page exists.
- `last_id` (String) Last spend alert ID in the page.
