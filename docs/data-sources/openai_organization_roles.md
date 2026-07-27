# openai_organization_roles

Lists OpenAI organization roles.

## Example Usage

```terraform
data "openai_organization_roles" "all" {
  limit = 100
}
```

## Schema

### Optional

- `after`, `limit`, `order` pagination and sort controls.

### Read-Only

- `items` (List) Organization roles.
- `has_more` (Boolean) Whether another page exists.
- `next` (String) Next cursor.
