# openai_organization_user_role

API group: [Roles and Assignments](../api-groups/roles-and-assignments.md).

Assigns an OpenAI organization role to a user. Destroy unassigns the role.

## Example Usage

```terraform
resource "openai_organization_user_role" "alice_auditor" {
  user_id = data.openai_organization_user.alice.id
  role_id = openai_organization_role.auditor.id
}
```

## Import

```sh
terraform import openai_organization_user_role.alice_auditor user_123/role_123
```

## Schema

### Required

- `user_id` (String) Organization user ID. Changes replace the assignment.
- `role_id` (String) Role ID. Changes replace the assignment.

### Read-Only

- `id` (String) Composite `user_id/role_id` ID.
- `name`, `description`, `resource_type`, `created_by` (String) Role metadata.
- `permissions` (Set of String) Permissions granted by the assigned role.
- `predefined_role` (Boolean) Whether OpenAI predefines the role.
- `created_at`, `updated_at` (Number) Role timestamps when returned.
- `assignment_sources` (List) Inherited assignment sources when returned.
