# IAC feasibility and community signal — 2026-09-24

## Evidence-linked gap matrix

| Rank | Control | Evidence and decision |
| --- | --- | --- |
| 1 | Project controls | OpenAI’s official [project controls guide](https://developers.openai.com/api/docs/guides/terraform/project-controls) and [rate limits and spend guide](https://developers.openai.com/api/docs/guides/terraform/rate-limits-and-spend) describe the durable project-administration boundary. A May 2026 [OpenAI Developer Community budget-control request](https://community.openai.com/t/api-endpoint-to-manage-project-budgets/1380411) is historical demand only: its endpoint claims were corrected/disputed, so the API contract here comes **only** from the official guide. Project spend limits, alerts, rate limits, model permissions, hosted-tool permissions, and data retention are registered where that Admin API supports their lifecycle. |
| 2 | Project access graph | The official [service accounts guide](https://developers.openai.com/api/docs/guides/terraform/service-accounts) supports project-scoped administrative identity management. Project users, groups, and user/group role assignments are registered where API lifecycle support exists. The [Terraform Reddit release announcement](https://www.reddit.com/r/Terraform/comments/1pha3k4/) is a community signal for roles/groups management, not evidence of an OpenAI feature request. |
| 3 | Key-state safety | [openai/terraform-provider-openai issue #87](https://github.com/openai/terraform-provider-openai/issues/87) requests write-only key handling. No new conventional-key design is added in this pass; existing explicitly supported admin and service-account key resources remain stateful secret material and require encrypted, access-controlled state. |
| 4 | Service tier and EKM exclusions | [jianyuan issue #255](https://github.com/jianyuan/terraform-provider-openai/issues/255) concerns service tier. [mkdev-me issues #33](https://github.com/mkdev-me/terraform-provider-openai/issues/33) and [#34](https://github.com/mkdev-me/terraform-provider-openai/issues/34) concern EKM. These are excluded because this pass does not have a stable documented Admin API lifecycle suitable for Terraform state. |
| 5 | Reliability and remaining candidates | [mkdev-me issue #11](https://github.com/mkdev-me/terraform-provider-openai/issues/11) concerns rate-limit/retry reliability, supporting explicit API error handling and delete confirmation. Project certificates and usage/cost reporting remain candidates, not registered surfaces. |

## Exclusions and interpretation

- **Undocumented service-tier/EKM controls:** excluded; the cited provider issues show demand but do not themselves establish a supported OpenAI Admin API contract.
- **Inference/file APIs:** excluded as application/workload operations rather than durable administration controls.
- **Project certificates and usage reporting:** retained as candidates pending lifecycle and schema review.
- **State consistency:** the [Terraform Reddit discussion](https://www.reddit.com/r/Terraform/comments/1loxsww/) is treated only as a community discussion about provider state consistency, not as a per-user-budget request.

This matrix ranks documented feasibility and community signals; availability still depends on account eligibility, permissions, and the OpenAI service tier.
