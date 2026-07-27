# Organization Groups API group

Use the Organization Groups API group to create groups, read groups, and manage
explicit user membership in groups.

## Terraform surfaces

| Surface | Kind | Purpose | Detailed docs |
| --- | --- | --- | --- |
| `openai_organization_group` | Resource | Create and update organization groups. | `docs/resources/openai_organization_group.md` |
| `openai_organization_groups` | Data source | List organization groups. | `docs/data-sources/openai_organization_groups.md` |
| `openai_organization_group` | Data source | Read one organization group by ID. | `docs/data-sources/openai_organization_group.md` |
| `openai_organization_group_user` | Resource | Add a user to a group. | `docs/resources/openai_organization_group_user.md` |
| `openai_organization_group_users` | Data source | List users in a group. | `docs/data-sources/openai_organization_group_users.md` |
| `openai_organization_group_user` | Data source | Read one group member. | `docs/data-sources/openai_organization_group_user.md` |
| `openai_organization_group_role` | Resource | Assign an organization role to a group. | `docs/resources/openai_organization_group_role.md` |
| `openai_organization_group_roles` | Data source | List role assignments for a group. | `docs/data-sources/openai_organization_group_roles.md` |
| `openai_organization_group_role` | Data source | Read one role assignment for a group. | `docs/data-sources/openai_organization_group_role.md` |

## Common workflow

```hcl
data "openai_organization_user" "alice" {
  id = "user_123"
}

resource "openai_organization_group" "engineering" {
  name = "Engineering"
}

resource "openai_organization_group_user" "engineering_alice" {
  group_id = openai_organization_group.engineering.id
  user_id  = data.openai_organization_user.alice.id
}

resource "openai_organization_group_role" "engineering_auditor" {
  group_id = openai_organization_group.engineering.id
  role_id  = openai_organization_role.auditor.id
}
```

## Import and destroy behavior

| Resource | Import ID | Destroy behavior |
| --- | --- | --- |
| `openai_organization_group` | `group_123` | Deletes the group. |
| `openai_organization_group_user` | `group_123/user_123` | Removes the user from the group. |
| `openai_organization_group_role` | `group_123/role_123` | Removes the role assignment from the group. |

## Operational notes

- Import existing groups before managing their membership in Terraform.
- For externally synchronized groups, check `is_scim_managed` in data sources before deciding whether Terraform should manage membership.
- Group membership destroy does not delete the user.
