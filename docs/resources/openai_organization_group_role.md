# openai_organization_group_role

API group: [Roles and Assignments](../api-groups/roles-and-assignments.md).

Assigns an OpenAI organization role to a group. Destroy unassigns the role.

## Example Usage

```terraform
resource "openai_organization_group_role" "engineering_auditor" {
  group_id = openai_organization_group.engineering.id
  role_id  = openai_organization_role.auditor.id
}
```

## Import

```sh
terraform import openai_organization_group_role.engineering_auditor group_123/role_123
```

## Schema

### Required

- `group_id` (String) Organization group ID. Changes replace the assignment.
- `role_id` (String) Role ID. Changes replace the assignment.

### Read-Only

- `id` (String) Composite `group_id/role_id` ID.
- `name`, `description`, `resource_type`, `created_by` (String) Role metadata.
- `permissions` (Set of String) Permissions granted by the assigned role.
- `predefined_role` (Boolean) Whether OpenAI predefines the role.
- `created_at`, `updated_at` (Number) Role timestamps when returned.
- `assignment_sources` (List) Inherited assignment sources when returned.
