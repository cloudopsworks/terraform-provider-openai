# openai_organization_invite

API group: [Organization Controls](../api-groups/organization-controls.md).

OpenAI API hierarchy: `Administration > Organization > Invites`.

Creates an OpenAI organization invite. Invites are immutable after creation;
changes to `email`, `role`, or `projects` replace the invite.

## Example Usage

```terraform
resource "openai_organization_invite" "platform_admin" {
  email = "platform-admin@example.com"
  role  = "owner"

  projects = [{
    id   = openai_project.platform.id
    role = "owner"
  }]
}
```

## Import

```sh
terraform import openai_organization_invite.platform_admin invite_123
```

## Schema

### Required

- `email` (String) Email address to invite. Changes replace the invite.
- `role` (String) Organization role for the invite. Supported values: `reader`, `owner`. Changes replace the invite.

### Optional

- `projects` (List) Project memberships granted when the invite is accepted. Omit to use OpenAI's default-project compatibility behavior, or set an empty list for no project membership. Changes replace the invite.
  - `id` (String) Project public ID.
  - `role` (String) Project role granted on invite acceptance. Supported values: `member`, `owner`.

### Read-Only

- `id` (String) OpenAI invite ID.
- `status` (String) Invite status such as `pending`, `accepted`, or `expired`.
- `created_at` (Number) Unix timestamp when the invite was sent.
- `accepted_at` (Number) Unix timestamp when the invite was accepted, when available.
- `expires_at` (Number) Unix timestamp when the invite expires, when available.

## Notes

- Destroy revokes/deletes the invite through the OpenAI API.
- Deleting an invite does not delete an accepted user.
