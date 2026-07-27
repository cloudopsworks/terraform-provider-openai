# openai_organization_group_roles

Lists organization role assignments for an OpenAI organization group.

## Example Usage

```terraform
data "openai_organization_group_roles" "engineering" {
  group_id = "group_123"
}
```

## Schema

### Required

- `group_id` (String) Group ID.

### Optional

- `after`, `limit`, `order` pagination and sort controls.

### Read-Only

- `items` (List) Group role assignments.
- `has_more` (Boolean) Whether another page exists.
- `next` (String) Next cursor.
