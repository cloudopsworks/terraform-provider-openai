# openai_organization_groups

API group: [Organization Groups](../api-groups/organization-groups.md).

OpenAI API hierarchy: `Administration > Organization > Groups`.

Lists OpenAI organization groups.

## Example Usage

```terraform
data "openai_organization_groups" "all" {
  limit = 100
}
```

## Schema

### Optional

- `after`, `limit`, `order` pagination and sort controls.

### Read-Only

- `items` (List) Organization groups.
- `has_more` (Boolean) Whether another page exists.
- `next` (String) Next cursor.
