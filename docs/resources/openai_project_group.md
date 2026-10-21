# openai_project_group

API group: [Projects](../api-groups/projects.md).

OpenAI API hierarchy: `Administration > Organization > Projects > Groups`.

Create a group membership in a project.

## Lifecycle and import safety

The membership role is bootstrap/create-only. Changing it replaces the membership,
which revokes group access before recreating it and may affect independently managed
group-role assignments. Review the replacement plan and use `prevent_destroy` where
that interruption is unacceptable.

Import is intentionally rejected: OpenAI omits the create-only membership role from
read/list responses, so import would produce incomplete state and could revoke access
on a later apply. Declare the membership and its role instead, or leave an existing
membership unmanaged.

## Notes

- Supply the OpenAI project ID using `project_id`.
