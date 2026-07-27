# openai_organization_group_role

Looks up one OpenAI organization role assignment for a group.

## Example Usage

```terraform
data "openai_organization_group_role" "engineering_auditor" {
  group_id = "group_123"
  role_id  = "role_123"
}
```

## Schema

### Required

- `group_id` (String) Group ID.
- `role_id` (String) Role ID.

### Read-Only

- Role metadata, permissions, timestamps, and assignment sources.
