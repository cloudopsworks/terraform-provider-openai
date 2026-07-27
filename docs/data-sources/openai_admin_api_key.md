# openai_admin_api_key

API group: [Admin API Keys](../api-groups/admin-api-keys.md).

Looks up one OpenAI organization admin API key by ID. The unredacted key value is not returned.

## Example Usage

```terraform
data "openai_admin_api_key" "automation" {
  id = "admin_key_123"
}
```

## Schema

### Required

- `id` (String) Admin API key ID.

### Read-Only

- `name`, `redacted_value`, `owner_type`, `owner_id`, `owner_name`, `owner_role` (String) Key metadata.
- `created_at`, `expires_at`, `last_used_at` (Number) Unix timestamps when returned.
