# Administration > Organization > Projects

Use the Projects API group for project lifecycle, access control, service accounts,
project-scoped API keys and custom roles, and project-level controls.

## OpenAI API hierarchy covered

- Administration > Organization > Projects
- Administration > Organization > Projects > Service Accounts > API Keys
- Administration > Organization > Projects > Roles
- Administration > Organization > Projects > Users
- Administration > Organization > Projects > Users > Roles
- Administration > Organization > Projects > Groups
- Administration > Organization > Projects > Groups > Roles
- Administration > Organization > Projects > Data Retention, Spend Limit, Spend Alerts, Rate Limits
- Administration > Organization > Projects > Model Permissions, Hosted Tool Permissions

## Terraform surfaces

| Surface | Kind | Purpose | Detailed docs |
| --- | --- | --- | --- |
| `openai_project`, `openai_projects` | Resource/data sources | Manage, read, or list projects. | `docs/resources/openai_project.md` |
| `openai_service_account`, `openai_service_accounts` | Resource/data sources | Manage, read, or list service accounts. | `docs/resources/openai_service_account.md` |
| `openai_project_api_key` | Resource | Create project service-account API keys. | `docs/resources/openai_project_api_key.md` |
| `openai_project_role`, `openai_project_roles` | Resource/data sources | Manage, read, or list project custom roles. | `docs/resources/openai_project_role.md` |
| `openai_project_user`, `openai_project_users` | Resource/data sources | Manage, read, or list project-user memberships. | `docs/resources/openai_project_user.md` |
| `openai_project_user_role`, `openai_project_user_roles` | Resource/data sources | Manage, read, or list project-user role assignments. | `docs/resources/openai_project_user_role.md` |
| `openai_project_group`, `openai_project_groups` | Resource/data sources | Manage, read, or list project-group memberships. | `docs/resources/openai_project_group.md` |
| `openai_project_group_role`, `openai_project_group_roles` | Resource/data sources | Manage, read, or list project-group role assignments. | `docs/resources/openai_project_group_role.md` |
| `openai_project_data_retention`, `openai_project_spend_limit` | Resource/data sources | Manage or read singleton project controls. | `docs/resources/openai_project_data_retention.md` |
| `openai_project_spend_alert`, `openai_project_spend_alerts` | Resource/data sources | Manage, read, or list spend alerts. | `docs/resources/openai_project_spend_alert.md` |
| `openai_project_rate_limit`, `openai_project_rate_limits` | Resource/data sources | Manage, read, or list rate limits. | `docs/resources/openai_project_rate_limit.md` |
| `openai_project_model_permissions`, `openai_project_hosted_tool_permissions` | Resource/data sources | Manage or read model and hosted-tool controls. | `docs/resources/openai_project_model_permissions.md` |

## Lifecycle and safety notes

- `openai_project_rate_limit`, `openai_project_hosted_tool_permissions`, and
  `openai_project_data_retention` have state-only destroy behavior: removing the
  Terraform resource leaves the remote setting unchanged. Use `prevent_destroy`
  for controls that must not be accidentally unmanaged. To roll back data retention
  before removal, explicitly apply `type = "organization_default"`, then remove the
  resource; apply the desired explicit value similarly for the other controls.
- `openai_project_group` is a bootstrap/create-only membership surface. Changing its
  role replaces the membership, revoking group access before recreating it; that can
  affect independently managed group roles. Import is intentionally rejected because
  OpenAI omits the create-only membership role from read/list responses.
- Project user/group role assignments are immutable. Change their project, principal,
  or role IDs by replacement and review the resulting access revoke/regrant plan.
