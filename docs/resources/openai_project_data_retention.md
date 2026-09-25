# openai_project_data_retention

API group: [Projects](../api-groups/projects.md).

OpenAI API hierarchy: `Administration > Organization > Projects > Data Retention`.

Manage a project data-retention policy.

## Lifecycle

OpenAI has retrieve and update operations but no delete or reset endpoint. Destroy
removes only Terraform state; the remote retention setting remains unchanged. Use
`lifecycle { prevent_destroy = true }` where this control must remain managed. To
roll back before removal, explicitly apply `type = "organization_default"`, then
remove the resource after accepting that the remote setting persists.

## Notes

- Supply the OpenAI project ID using `project_id`.
- Import the singleton with its project ID.
