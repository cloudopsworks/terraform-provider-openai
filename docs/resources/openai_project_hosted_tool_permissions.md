# openai_project_hosted_tool_permissions

API group: [Projects](../api-groups/projects.md).

OpenAI API hierarchy: `Administration > Organization > Projects > Hosted Tool Permissions`.

Manage project hosted-tool permissions.

## Lifecycle

OpenAI has retrieve and update operations but no delete or reset endpoint. Destroy
removes only Terraform state; remote hosted-tool settings remain unchanged. Use
`lifecycle { prevent_destroy = true }` for safety-critical permissions. To roll back,
explicitly apply the intended enabled/disabled values before removing this resource,
then remove it only after accepting that those remote values persist.

## Notes

- Supply the OpenAI project ID using `project_id`.
- Import the singleton with its project ID.
