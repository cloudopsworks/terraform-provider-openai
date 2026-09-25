# OpenAI Administration API hierarchy

This directory mirrors the OpenAI Administration API documentation hierarchy and
maps each documented OpenAI node to the Terraform provider surfaces implemented
in this repository. Use these guides with the generated Terraform resource and
data-source reference pages under `docs/resources/` and `docs/data-sources/`.

OpenAI API hierarchy represented here:

- Administration
  - Organization
    - Admin API Keys
    - Invites
    - Users
      - Roles
    - Groups
      - Users
      - Roles
    - Roles
    - Data Retention
    - Spend Limit
    - Spend Alerts
    - Certificates
    - Projects
      - Service Accounts
        - API Keys
      - API Keys
      - Roles
      - Users
        - Roles
      - Groups
        - Roles
      - Rate Limits
      - Model Permissions
      - Hosted Tool Permissions
      - Data Retention
      - Spend Limit
      - Spend Alerts
      - Certificates *(evaluated; not implemented yet)*

## Implemented Terraform mapping

| OpenAI API hierarchy | Guide | Primary Terraform surfaces |
| --- | --- | --- |
| Administration > Organization > Admin API Keys | `docs/api-groups/admin-api-keys.md` | `openai_admin_api_key`, `openai_admin_api_keys` |
| Administration > Organization > Invites | `docs/api-groups/organization-controls.md` | `openai_organization_invite`, `openai_organization_invites` |
| Administration > Organization > Users | `docs/api-groups/organization-users.md` | `openai_organization_user`, `openai_organization_users` |
| Administration > Organization > Users > Roles | `docs/api-groups/roles-and-assignments.md` | `openai_organization_user_role`, `openai_organization_user_roles` |
| Administration > Organization > Groups | `docs/api-groups/organization-groups.md` | `openai_organization_group`, `openai_organization_groups` |
| Administration > Organization > Groups > Users | `docs/api-groups/organization-groups.md` | `openai_organization_group_user`, `openai_organization_group_users` |
| Administration > Organization > Groups > Roles | `docs/api-groups/organization-groups.md` | `openai_organization_group_role`, `openai_organization_group_roles` |
| Administration > Organization > Roles | `docs/api-groups/roles-and-assignments.md` | `openai_organization_role`, `openai_organization_roles` |
| Administration > Organization > Data Retention | `docs/api-groups/organization-controls.md` | `openai_organization_data_retention` |
| Administration > Organization > Spend Limit | `docs/api-groups/organization-controls.md` | `openai_organization_spend_limit` |
| Administration > Organization > Spend Alerts | `docs/api-groups/organization-controls.md` | `openai_organization_spend_alert`, `openai_organization_spend_alerts` |
| Administration > Organization > Certificates | `docs/api-groups/organization-controls.md` | `openai_organization_certificate`, `openai_organization_certificates` |
| Administration > Organization > Projects | `docs/api-groups/projects.md` | `openai_project`, `openai_projects` |
| Administration > Organization > Projects > Service Accounts | `docs/api-groups/projects.md` | `openai_service_account`, `openai_service_accounts` |
| Administration > Organization > Projects > Service Accounts > API Keys | `docs/api-groups/projects.md` | `openai_project_api_key` |
| Administration > Organization > Projects > Roles | `docs/api-groups/projects.md`, `docs/api-groups/roles-and-assignments.md` | `openai_project_role`, `openai_project_roles` |
| Administration > Organization > Projects > Users | `docs/api-groups/projects.md` | `openai_project_user`, `openai_project_users` |
| Administration > Organization > Projects > Users > Roles | `docs/api-groups/projects.md` | `openai_project_user_role`, `openai_project_user_roles` |
| Administration > Organization > Projects > Groups | `docs/api-groups/projects.md` | `openai_project_group`, `openai_project_groups` |
| Administration > Organization > Projects > Groups > Roles | `docs/api-groups/projects.md` | `openai_project_group_role`, `openai_project_group_roles` |
| Administration > Organization > Projects > Data Retention | `docs/api-groups/projects.md` | `openai_project_data_retention` |
| Administration > Organization > Projects > Spend Limit | `docs/api-groups/projects.md` | `openai_project_spend_limit` |
| Administration > Organization > Projects > Spend Alerts | `docs/api-groups/projects.md` | `openai_project_spend_alert`, `openai_project_spend_alerts` |
| Administration > Organization > Projects > Rate Limits | `docs/api-groups/projects.md` | `openai_project_rate_limit`, `openai_project_rate_limits` |
| Administration > Organization > Projects > Model Permissions | `docs/api-groups/projects.md` | `openai_project_model_permissions` |
| Administration > Organization > Projects > Hosted Tool Permissions | `docs/api-groups/projects.md` | `openai_project_hosted_tool_permissions` |

## Remaining project candidates

Project certificates and usage/cost reporting remain candidates for later evaluation; this provider does not currently register Terraform surfaces for them.
