# openai_project_roles

API group: [Roles and Assignments](../api-groups/roles-and-assignments.md).

Lists OpenAI project roles.

## Example Usage

```terraform
data "openai_project_roles" "app" {
  project_id = "proj_123"
  limit      = 100
}
```

## Schema

### Required

- `project_id` (String) Project ID.

### Optional

- `after`, `limit`, `order` pagination and sort controls.

### Read-Only

- `items` (List) Project roles.
- `has_more` (Boolean) Whether another page exists.
- `next` (String) Next cursor.
