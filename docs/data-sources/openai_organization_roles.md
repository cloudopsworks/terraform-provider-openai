# openai_organization_roles

API group: [Roles and Assignments](../api-groups/roles-and-assignments.md).

OpenAI API hierarchy: `Administration > Organization > Roles`.

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
