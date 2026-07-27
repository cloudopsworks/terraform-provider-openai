# Roles and Assignments API group

Use the Roles and Assignments API group to manage organization custom roles,
project custom roles, and organization role assignments for users or groups.

## Terraform surfaces

| Surface | Kind | Purpose | Detailed docs |
| --- | --- | --- | --- |
| `openai_organization_role` | Resource | Create and update organization-scoped custom roles. | `docs/resources/openai_organization_role.md` |
| `openai_organization_roles` | Data source | List organization-scoped roles. | `docs/data-sources/openai_organization_roles.md` |
| `openai_organization_role` | Data source | Read one organization-scoped role. | `docs/data-sources/openai_organization_role.md` |
| `openai_project_role` | Resource | Create and update project-scoped custom roles. | `docs/resources/openai_project_role.md` |
| `openai_project_roles` | Data source | List project-scoped roles. | `docs/data-sources/openai_project_roles.md` |
| `openai_project_role` | Data source | Read one project-scoped role. | `docs/data-sources/openai_project_role.md` |
| `openai_organization_user_role` | Resource | Assign an organization role to a user. | `docs/resources/openai_organization_user_role.md` |
| `openai_organization_user_roles` | Data source | List organization role assignments for a user. | `docs/data-sources/openai_organization_user_roles.md` |
| `openai_organization_user_role` | Data source | Read one organization role assignment for a user. | `docs/data-sources/openai_organization_user_role.md` |
| `openai_organization_group_role` | Resource | Assign an organization role to a group. | `docs/resources/openai_organization_group_role.md` |
| `openai_organization_group_roles` | Data source | List organization role assignments for a group. | `docs/data-sources/openai_organization_group_roles.md` |
| `openai_organization_group_role` | Data source | Read one organization role assignment for a group. | `docs/data-sources/openai_organization_group_role.md` |

## Common workflow

```hcl
resource "openai_organization_role" "auditor" {
  name        = "auditor"
  description = "Read-only organization auditor"
  permissions = [
    "organization.users.read",
  ]
}

resource "openai_organization_group_role" "engineering_auditor" {
  group_id = openai_organization_group.engineering.id
  role_id  = openai_organization_role.auditor.id
}

resource "openai_project_role" "api_key_reader" {
  project_id  = openai_project.app.id
  name        = "api-key-reader"
  description = "Read project API keys"
  permissions = [
    "project.api_keys.read",
  ]
}
```

## Import and destroy behavior

| Resource | Import ID | Destroy behavior |
| --- | --- | --- |
| `openai_organization_role` | `role_123` | Deletes the organization role. |
| `openai_project_role` | `proj_123/role_123` | Deletes the project role. |
| `openai_organization_user_role` | `user_123/role_123` | Removes the assignment from the user. |
| `openai_organization_group_role` | `group_123/role_123` | Removes the assignment from the group. |

## Operational notes

- Prefer custom roles with the smallest permission set needed by an automation or team.
- Assign organization roles to groups when possible; use direct user assignments for exceptions.
- Import existing roles and assignments before Terraform manages or destroys them.
