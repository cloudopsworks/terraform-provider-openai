# openai_organization_group_user

Manages membership of an OpenAI organization user in a group. Destroy removes the user from the group.

## Example Usage

```terraform
resource "openai_organization_group_user" "engineering_alice" {
  group_id = openai_organization_group.engineering.id
  user_id  = data.openai_organization_user.alice.id
}
```

## Import

```sh
terraform import openai_organization_group_user.engineering_alice group_123/user_123
```

## Schema

### Required

- `group_id` (String) Group ID. Changes replace the membership.
- `user_id` (String) User ID. Changes replace the membership.

### Read-Only

- `id` (String) Composite `group_id/user_id` ID.
- `email`, `name`, `picture`, `user_type` (String) User details returned by OpenAI.
- `is_service_account` (Boolean) Whether the member is a service account.
