# Organization Users API group

Use the Organization Users API group to discover existing users before managing
group membership or role assignments. This provider exposes users as data sources
only; it does not invite, create, or delete organization users.

## Terraform surfaces

| Surface | Kind | Purpose | Detailed docs |
| --- | --- | --- | --- |
| `openai_organization_user` | Data source | Read one organization user by ID. | `docs/data-sources/openai_organization_user.md` |
| `openai_organization_users` | Data source | List users, optionally filtered by email addresses. | `docs/data-sources/openai_organization_users.md` |
| `openai_organization_user_role` | Resource | Assign an organization role to a user. | `docs/resources/openai_organization_user_role.md` |
| `openai_organization_user_roles` | Data source | List role assignments for a user. | `docs/data-sources/openai_organization_user_roles.md` |
| `openai_organization_user_role` | Data source | Read one role assignment for a user. | `docs/data-sources/openai_organization_user_role.md` |

## Common workflow

```hcl
data "openai_organization_user" "alice" {
  id = "user_123"
}

resource "openai_organization_role" "auditor" {
  name        = "auditor"
  description = "Read-only organization auditor"
  permissions = [
    "organization.users.read",
  ]
}

resource "openai_organization_user_role" "alice_auditor" {
  user_id = data.openai_organization_user.alice.id
  role_id = openai_organization_role.auditor.id
}
```

## Operational notes

- Keep user identity lifecycle outside this provider and import/reference users by stable ID.
- Prefer group-based assignments when multiple users need the same organization role.
- Destroying `openai_organization_user_role` removes only the role assignment; it does not delete the user.
