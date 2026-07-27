# openai_organization_group

API group: [Organization Groups](../api-groups/organization-groups.md).

OpenAI API hierarchy: `Administration > Organization > Groups`.

Manages an OpenAI organization group. Destroy deletes the group through the OpenAI Admin API.

## Example Usage

```terraform
resource "openai_organization_group" "engineering" {
  name = "Engineering"
}
```

## Import

```sh
terraform import openai_organization_group.engineering group_123
```

## Schema

### Required

- `name` (String) Group display name.

### Read-Only

- `id` (String) Group ID.
- `group_type` (String) Group type returned by OpenAI.
- `is_scim_managed` (Boolean) Whether SCIM manages the group.
- `created_at` (Number) Unix timestamp for group creation.
