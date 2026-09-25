# openai_project_spend_alert

API group: [Projects](../api-groups/projects.md).

OpenAI API hierarchy: `Administration > Organization > Projects > Spend Alerts`.

Manage one project spend alert.

## Notes

- Supply the OpenAI project ID using `project_id`.
- Import existing configuration before Terraform manages destructive operations.
- To clear an existing `notification_channel.subject_prefix` on update, omit the optional field. An empty string is rejected.
