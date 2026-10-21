# openai_project_rate_limit

API group: [Projects](../api-groups/projects.md).

OpenAI API hierarchy: `Administration > Organization > Projects > Rate Limits`.

Manage one project rate-limit configuration.

## Lifecycle

OpenAI exposes list and update operations, but no create, delete, or reset endpoint.
Destroy removes only Terraform state; the remote rate limit remains active. Use
`lifecycle { prevent_destroy = true }` for a limit that must remain managed. Before
removing this resource, explicitly apply the intended rollback values, then remove it
from configuration only after accepting that the remote values persist.

## Notes

- Supply the OpenAI project ID using `project_id`.
- Import an existing rate limit with `project_id/rate_limit_id`.
