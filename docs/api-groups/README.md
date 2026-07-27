# OpenAI Administration API groups

This directory contains task-oriented guides grouped by OpenAI Administration API
area. Use these guides with the generated Terraform resource and data-source
reference pages under `docs/resources/` and `docs/data-sources/`.

| API group | Guide | Primary Terraform surfaces |
| --- | --- | --- |
| Projects | `docs/api-groups/projects.md` | `openai_project`, `openai_service_account`, `openai_project_api_key`, `openai_project_role` |
| Admin API Keys | `docs/api-groups/admin-api-keys.md` | `openai_admin_api_key` and admin-key data sources |
| Organization Controls | `docs/api-groups/organization-controls.md` | `openai_organization_invite`, `openai_organization_data_retention`, `openai_organization_spend_limit`, `openai_organization_spend_alert`, `openai_organization_certificate` |
| Organization Users | `docs/api-groups/organization-users.md` | user lookup data sources and user role assignments |
| Organization Groups | `docs/api-groups/organization-groups.md` | group, group membership, and group role assignment surfaces |
| Roles and Assignments | `docs/api-groups/roles-and-assignments.md` | organization/project roles and user/group role assignments |
