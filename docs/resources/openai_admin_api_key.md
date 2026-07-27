# openai_admin_api_key

Manages an OpenAI organization admin API key. The unredacted `value` is returned only during create and is stored as Sensitive Terraform state. Destroy deletes the key through the OpenAI Admin API.

Use `lifecycle { prevent_destroy = true }` for break-glass or critical automation keys.

## Example Usage

```terraform
resource "openai_admin_api_key" "automation" {
  name            = "terraform-automation"
  expire_in_hours = 24
  scopes = [
    "organization.users.read",
    "organization.projects.read",
  ]

  lifecycle {
    prevent_destroy = true
  }
}
```

## Import

```sh
terraform import openai_admin_api_key.automation admin_key_123
```

## Schema

### Required

- `name` (String) Admin API key name. Changes replace the key.

### Optional

- `expires_in_seconds` (Number) Optional lifetime in seconds. Mutually exclusive with `expire_in_hours` and `expire_in_days`. Changes replace the key.
- `expire_in_hours` (Number) Optional lifetime in hours. Mutually exclusive with `expires_in_seconds` and `expire_in_days`. Changes replace the key.
- `expire_in_days` (Number) Optional lifetime in days. Mutually exclusive with `expires_in_seconds` and `expire_in_hours`. Changes replace the key.
- `scopes` (Set of String) Optional admin API key scopes. Values are sent as the create request `scopes` array. Changes replace the key.

### Read-Only

- `id` (String) Admin API key ID.
- `value` (String, Sensitive) Unredacted key value returned only on create.
- `redacted_value` (String) Redacted key value.
- `owner_type`, `owner_id`, `owner_name`, `owner_role` (String) Key owner details.
- `created_at`, `expires_at`, `last_used_at` (Number) Unix timestamps when returned.

## Notes

- Omit `expires_in_seconds`, `expire_in_hours`, and `expire_in_days` for a non-expiring key.
- Configure only one expiration input. The provider converts hours and days to OpenAI's `expires_in_seconds` create parameter.
- `scopes` is create-only for provider maintenance. Scope changes force replacement, and configured scopes are preserved in Terraform state across refresh because current OpenAI admin-key metadata does not expose scope details through the typed SDK model.
- Imported admin keys have unknown external scope configuration to Terraform. Leave `scopes` unset for imported keys unless you intend to replace them with a managed scope set.
- Destroy and replacement call the OpenAI revoke/delete API and now fail the Terraform operation if OpenAI does not return `deleted=true` for the requested key ID.
