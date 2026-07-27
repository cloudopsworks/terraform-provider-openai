# openai_organization_user_roles

API group: [Roles and Assignments](../api-groups/roles-and-assignments.md).

OpenAI API hierarchy: `Administration > Organization > Users > Roles`.

Lists organization role assignments for an OpenAI organization user.

## Example Usage

```terraform
data "openai_organization_user_roles" "alice" {
  user_id = "user_123"
}
```

## Schema

### Required

- `user_id` (String) User ID.

### Optional

- `after`, `limit`, `order` pagination and sort controls.

### Read-Only

- `items` (List) User role assignments.
- `has_more` (Boolean) Whether another page exists.
- `next` (String) Next cursor.
