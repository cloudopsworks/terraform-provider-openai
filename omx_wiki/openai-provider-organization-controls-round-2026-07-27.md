---
title: "OpenAI Provider Organization Controls Round 2026-07-27"
tags: ["terraform-provider", "openai", "admin-api", "organization-controls", "project-subapis", "documentation", "lessons-learned"]
created: 2026-07-27T18:09:03.353Z
updated: 2026-07-27T18:09:03.353Z
sources: []
links: ["openai-provider-progress-checkpoint-2026-07-27.md"]
category: session-log
confidence: medium
schemaVersion: 1
---

# OpenAI Provider Organization Controls Round 2026-07-27

# OpenAI Provider Organization Controls Round 2026-07-27

## Scope

Implemented the next OpenAI Administration API group for organization-wide controls and evaluated the project-scoped sub-APIs for maintainability opportunities. This page follows [[openai-provider-progress-checkpoint-2026-07-27]].

## Implemented provider surfaces

Organization controls now include these resources:

- `openai_organization_invite` for creating and revoking pending invites. Email, organization role, and project grants are immutable and force replacement.
- `openai_organization_data_retention` as a singleton resource using synthetic ID `organization`. Create/update call the update endpoint; destroy is state-only with a warning because OpenAI exposes no delete/reset endpoint.
- `openai_organization_spend_limit` as a singleton hard monthly spend limit. Destroy calls the OpenAI delete endpoint.
- `openai_organization_spend_alert` for email spend alerts. Recipient sets are sorted before state writes for stable plans.
- `openai_organization_certificate` for upload, name update, organization activation/deactivation, and delete. Active certificates are deactivated before deletion.

Organization controls now include these data sources:

- `openai_organization_invite`, `openai_organization_invites`
- `openai_organization_data_retention`
- `openai_organization_spend_limit`
- `openai_organization_spend_alert`, `openai_organization_spend_alerts`
- `openai_organization_certificate`, `openai_organization_certificates`

Provider registry totals moved from 10 resources / 20 data sources to 15 resources / 28 data sources.

## Client-layer changes

`internal/client` now normalizes SDK models for invites, data retention, spend limits, spend alerts, and certificates. Delete methods validate deletion confirmations where the OpenAI SDK/API returns them. Certificate read fills `active` by combining specific certificate retrieval with certificate-list activation metadata.

## Project sub-API evaluation

The OpenAI Go SDK exposes project-scoped sub-APIs for users, user roles, groups, group roles, model permissions, hosted-tool permissions, rate limits, data retention, spend limits, spend alerts, and certificate activation/deactivation. Current provider coverage already includes projects, service accounts, service-account API key creation, project API keys, and project roles.

Recommended next architecture work before implementing the remaining project sub-APIs:

1. Introduce shared cursor pagination structs and mappers for `after`, `before`, `limit`, `order`, `has_more`, `last_id`, and `next`.
2. Centralize import-ID parsing/formatting for singleton, two-part, and three-part resource IDs.
3. Generalize singleton-setting resources for organization/project data retention, spend limits, model permissions, and hosted-tool permissions.
4. Generalize role assignment and membership resources with scope/principal descriptors instead of duplicating user/group/project combinations.
5. Reuse certificate details/item models for organization upload plus organization/project activation state.

## Documentation added

- `docs/api-groups/organization-controls.md`
- Resource docs for all five new resources.
- Data-source docs for all eight new data sources.
- Example Terraform snippets under `examples/resources/` and `examples/data-sources/`.
- README.yaml updates for generated README usage and examples.

## Lessons learned

- OpenAI organization certificates require explicit activation after upload; Terraform should model desired activation separately from upload.
- Some organization settings are true singletons; synthetic IDs are preferable to inventing fake API object IDs.
- Data-retention destroy semantics must be conservative and explicit because there is no delete/reset endpoint.
- Spend limit and spend alert share amount/currency/interval concepts, but alerts additionally need a reusable notification-channel model.
- Project sub-API expansion should happen after extracting shared scoped CRUD/list/singleton helpers to avoid duplicating the new organization-control patterns.
