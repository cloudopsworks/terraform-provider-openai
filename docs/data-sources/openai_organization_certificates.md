# openai_organization_certificates

API group: [Organization Controls](../api-groups/organization-controls.md).

Lists OpenAI organization certificates and active status. PEM content is not
included in list results.

## Example Usage

```terraform
data "openai_organization_certificates" "all" {
  limit = 100
  order = "desc"
}
```

## Schema

### Optional

- `after` (String) Pagination cursor.
- `limit` (Number) Maximum certificates to return.
- `order` (String) Sort order by creation time: `asc` or `desc`.

### Read-Only

- `items` (List) Certificate records.
- `has_more` (Boolean) Whether another page exists.
- `last_id` (String) Last certificate ID in the page.
