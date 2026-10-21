# Administration > Organization > Controls

Use the Organization Controls API group to manage OpenAI organization-wide
administration controls that are not tied to a single project: invites, data
retention, hard spend limits, spend alerts, and certificate upload/activation.

## OpenAI API hierarchy covered

- Administration > Organization > Invites
- Administration > Organization > Data Retention
- Administration > Organization > Spend Limit
- Administration > Organization > Spend Alerts
- Administration > Organization > Certificates

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

## Project sub-API status and maintainability opportunities

The prior evaluation is now partly implemented. This table is the current status;
project certificates and usage/cost reporting remain candidates.

| Project sub-API | Current provider coverage | Notes |
| --- | --- | --- |
| Project users and groups | `openai_project_user(s)`, `openai_project_group(s)` | Project-group import is intentionally unavailable: read/list omit the create-only role, so import cannot safely hydrate a no-op state. |
| Project user and group roles | `openai_project_user_role(s)`, `openai_project_group_role(s)` | Scoped assignment helpers preserve the project/principal/role identity. |
| Project policies | Project data retention, spend limit, spend alerts, rate limits, model permissions, and hosted-tool permissions | Singleton and targeted policy surfaces are registered under the Projects API group. |
| Project certificates and usage | Not implemented | Retained as candidates pending documented lifecycle/API review. |

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
