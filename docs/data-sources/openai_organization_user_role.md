# openai_organization_user_role

API group: [Roles and Assignments](../api-groups/roles-and-assignments.md).

Looks up one OpenAI organization role assignment for a user.

## Example Usage

```terraform
data "openai_organization_user_role" "alice_auditor" {
  user_id = "user_123"
  role_id = "role_123"
}
```

## Schema

### Required

- `user_id` (String) User ID.
- `role_id` (String) Role ID.

### Read-Only

- Role metadata, permissions, timestamps, and assignment sources.
