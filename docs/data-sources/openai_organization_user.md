# openai_organization_user

API group: [Organization Users](../api-groups/organization-users.md).

OpenAI API hierarchy: `Administration > Organization > Users`.

Looks up one OpenAI organization user by ID. The current OpenAI Go SDK exposes user read/list/update/delete but not create, so this provider exposes users as data sources and manages access through groups and role assignments.

## Example Usage

```terraform
data "openai_organization_user" "alice" {
  id = "user_123"
}
```

## Schema

### Required

- `id` (String) Organization user ID.

### Read-Only

- User profile, status, role, project, SCIM, and timestamp attributes returned by OpenAI.
