# openai_organization_invite

API group: [Organization Controls](../api-groups/organization-controls.md).

Reads one OpenAI organization invite by ID.

## Example Usage

```terraform
data "openai_organization_invite" "pending" {
  id = "invite_123"
}
```

## Schema

### Required

- `id` (String) OpenAI invite ID.

### Read-Only

- `email` (String) Email address that received the invite.
- `role` (String) Organization role granted by the invite.
- `projects` (List) Project memberships granted when the invite is accepted.
- `status` (String) Invite status.
- `created_at` (Number) Unix timestamp when the invite was sent.
- `accepted_at` (Number) Unix timestamp when the invite was accepted, when available.
- `expires_at` (Number) Unix timestamp when the invite expires, when available.
