# openai_admin_api_key

Manages an OpenAI organization admin API key. The unredacted `value` is returned only during create and is stored as Sensitive Terraform state. Destroy deletes the key through the OpenAI Admin API.

Use `lifecycle { prevent_destroy = true }` for break-glass or critical automation keys.

## Example Usage

```terraform
resource "openai_admin_api_key" "automation" {
  name = "terraform-automation"

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

- `expires_in_seconds` (Number) Optional lifetime in seconds. Changes replace the key.

### Read-Only

- `id` (String) Admin API key ID.
- `value` (String, Sensitive) Unredacted key value returned only on create.
- `redacted_value` (String) Redacted key value.
- `owner_type`, `owner_id`, `owner_name`, `owner_role` (String) Key owner details.
- `created_at`, `expires_at`, `last_used_at` (Number) Unix timestamps when returned.
