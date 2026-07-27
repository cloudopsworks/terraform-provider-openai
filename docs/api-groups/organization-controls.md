# Organization Controls API group

Use the Organization Controls API group to manage OpenAI organization-wide
administration controls that are not tied to a single project: invites, data
retention, hard spend limits, spend alerts, and certificate upload/activation.

## Terraform surfaces

| Surface | Kind | Purpose | Detailed docs |
| --- | --- | --- | --- |
| `openai_organization_invite` | Resource | Create and revoke organization invites. | `docs/resources/openai_organization_invite.md` |
| `openai_organization_invites` | Data source | List organization invites. | `docs/data-sources/openai_organization_invites.md` |
| `openai_organization_invite` | Data source | Read one organization invite by ID. | `docs/data-sources/openai_organization_invite.md` |
| `openai_organization_data_retention` | Resource | Manage the organization data-retention mode. | `docs/resources/openai_organization_data_retention.md` |
| `openai_organization_data_retention` | Data source | Read the current organization data-retention mode. | `docs/data-sources/openai_organization_data_retention.md` |
| `openai_organization_spend_limit` | Resource | Manage the organization hard monthly spend limit. | `docs/resources/openai_organization_spend_limit.md` |
| `openai_organization_spend_limit` | Data source | Read the current organization hard spend limit. | `docs/data-sources/openai_organization_spend_limit.md` |
| `openai_organization_spend_alert` | Resource | Create, update, and delete organization spend alerts. | `docs/resources/openai_organization_spend_alert.md` |
| `openai_organization_spend_alerts` | Data source | List organization spend alerts. | `docs/data-sources/openai_organization_spend_alerts.md` |
| `openai_organization_spend_alert` | Data source | Read one organization spend alert by ID. | `docs/data-sources/openai_organization_spend_alert.md` |
| `openai_organization_certificate` | Resource | Upload organization certificates and activate/deactivate them. | `docs/resources/openai_organization_certificate.md` |
| `openai_organization_certificates` | Data source | List organization certificates and active status. | `docs/data-sources/openai_organization_certificates.md` |
| `openai_organization_certificate` | Data source | Read one organization certificate by ID, optionally including PEM content. | `docs/data-sources/openai_organization_certificate.md` |

## Common workflow

```hcl
resource "openai_organization_invite" "platform_admin" {
  email = "platform-admin@example.com"
  role  = "owner"

  projects = [{
    id   = openai_project.platform.id
    role = "owner"
  }]
}

resource "openai_organization_data_retention" "default" {
  type = "modified_abuse_monitoring"
}

resource "openai_organization_spend_limit" "monthly" {
  threshold_amount = 250000 # USD cents
}

resource "openai_organization_spend_alert" "eighty_percent" {
  threshold_amount = 200000 # USD cents

  notification_channel = {
    recipients     = ["platform-finops@example.com"]
    subject_prefix = "[OpenAI spend]"
  }
}

resource "openai_organization_certificate" "egress_root" {
  name        = "egress-root-ca"
  certificate = file("${path.module}/egress-root-ca.pem")
  active      = true
}
```

## Singleton controls

`openai_organization_data_retention` and `openai_organization_spend_limit` are
single organization-level settings. They use the synthetic Terraform ID
`organization` because the OpenAI API exposes one setting for the current
organization rather than an independent object ID.

| Resource | Create/Update behavior | Destroy behavior |
| --- | --- | --- |
| `openai_organization_data_retention` | Calls the OpenAI data-retention update endpoint. | Removes Terraform state only and emits a warning; OpenAI exposes retrieve/update but no delete/reset endpoint. |
| `openai_organization_spend_limit` | Calls the OpenAI spend-limit update endpoint. | Deletes the organization spend limit through the OpenAI API. |

## Invites

Invites are immutable after creation. Changes to `email`, `role`, or `projects`
replace the invite. Deleting the resource revokes/deletes the pending invite;
accepted users remain organization users and should be managed through groups or
role assignments.

`projects` is optional and create-only:

- Omit it to let OpenAI apply its default-project compatibility behavior.
- Set it to an empty list to invite the user with no project membership.
- Set one or more `{ id, role }` objects to grant project access on acceptance.

## Spend controls

Amounts are expressed in cents. This provider currently documents and validates
OpenAI's current organization-level values:

| Field | Supported value |
| --- | --- |
| `currency` | `USD` |
| `interval` | `month` |

Use the spend-limit resource for a hard cap and spend-alert resources for email
notifications before the cap is reached.

## Certificates

Organization certificate upload does not automatically activate the certificate.
Set `active = true` on `openai_organization_certificate` to activate it after
upload. Destroy deactivates an active certificate first, then deletes it.

Certificate contents are sensitive:

- `certificate` on the resource is required for create and is never returned by
  normal refresh.
- `certificate_details.content` is populated only by the single-certificate data
  source when `include_content = true` and is stored as Sensitive state.
- List data sources intentionally omit PEM content.

## Project sub-API evaluation and maintainability opportunities

The OpenAI Go SDK exposes project-scoped administration sub-APIs in addition to
the organization-wide controls implemented here. Current coverage and suggested
next work:

| Project sub-API | Current provider coverage | Recommended next step / generalization opportunity |
| --- | --- | --- |
| Project lifecycle | `openai_project`, `openai_projects`, `openai_project` | Continue using the existing project resource/data-source pair; factor singleton ID parsing helpers into a shared import-ID module. |
| Project users | Not implemented as first-class project membership resources in this round. | Add `openai_project_user` resource/data source and `openai_project_users` list data source. Reuse the organization group-user membership pattern with `project_id/user_id` import IDs. |
| Project user roles | Not implemented in this round. | Generalize the existing organization user/group role-assignment helpers to accept a scope descriptor (`organization`, `project`) and principal descriptor (`user`, `group`). |
| Project groups | Not implemented in this round. | Mirror `openai_organization_group_user` as project group membership, with shared add/read/list/delete membership plumbing. |
| Project group roles | Not implemented in this round. | Reuse the same scoped role-assignment abstraction as project user roles. |
| Project roles | `openai_project_role`, `openai_project_roles`, `openai_project_role` | Keep existing surface; extract common role create/update/list/delete mapping shared with organization roles. |
| Service accounts | `openai_service_account`, `openai_service_accounts`, `openai_service_account` | Already uses shared API-key creation when service-account scopes require a scoped bootstrap key. Continue factoring project/service-account/key import ID parsing. |
| Project API keys | `openai_project_api_key` plus project key list/read through service-account context | Extend read/list coverage for project-level key inventory if OpenAI exposes owner metadata needed by operators. Keep create-only scopes in shared service-account API-key code. |
| Model permissions | Not implemented in this round. | Add singleton resource/data source using the same create/update-as-upsert pattern as data retention, with explicit delete/reset semantics from the SDK. |
| Hosted-tool permissions | Not implemented in this round. | Add singleton resource/data source; no delete endpoint exists in the current SDK, so document state-only destroy if exposed as a resource. |
| Rate limits | Not implemented in this round. | Add data source for model/rate-limit discovery and targeted resource for per-rate-limit update. Reuse pagination and `project_id/rate_limit_id` import parsing. |
| Project data retention | Not implemented in this round. | Generalize organization data-retention models to a scoped singleton helper keyed by `project_id`. |
| Project spend limit | Not implemented in this round. | Generalize organization spend-limit models to a scoped singleton helper keyed by `project_id`; keep delete-confirmation logic shared. |
| Project spend alerts | Not implemented in this round. | Reuse organization spend-alert notification-channel and CRUD/list mapping with `project_id` added to IDs and requests. |
| Project certificates | Not implemented in this round. | Reuse organization certificate item/details models for project certificate activation/deactivation lists; certificate upload remains organization-level. |

The strongest architectural cleanup opportunity is a small internal framework for
scoped Admin API surfaces:

1. Shared pagination request/response structs for cursor pages (`after`, `before`,
   `limit`, `order`, `has_more`, `last_id`, `next`).
2. Shared import-ID parsing and formatter helpers for singleton, two-part, and
   three-part IDs.
3. Shared singleton-setting resource shape for organization/project data
   retention, model permissions, hosted-tool permissions, and spend limits.
4. Shared role-assignment and membership adapters for user/group principals across
   organization and project scopes.
5. Shared certificate models for organization upload and project/organization
   activation state.

Those refactors should be done before adding the remaining project sub-APIs so
new resources are smaller and regression tests can focus on endpoint-specific
mapping and lifecycle semantics.
