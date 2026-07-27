# Administration > Organization > Admin API Keys

Use the Admin API Keys API group to create, rotate, read, and revoke organization
admin API keys used by automation that calls OpenAI Administration APIs.

## OpenAI API hierarchy covered

- Administration > Organization > Admin API Keys

## Terraform surfaces

| Surface | Kind | Purpose | Detailed docs |
| --- | --- | --- | --- |
| `openai_admin_api_key` | Resource | Issue organization admin API keys with optional expiration. | `docs/resources/openai_admin_api_key.md` |
| `openai_admin_api_keys` | Data source | List admin API key metadata. Values are redacted. | `docs/data-sources/openai_admin_api_keys.md` |
| `openai_admin_api_key` | Data source | Read one admin API key metadata record by ID. | `docs/data-sources/openai_admin_api_key.md` |

## Common workflow

```hcl
resource "openai_admin_api_key" "automation" {
  name           = "terraform-admin-automation"
  expire_in_days = 30

  lifecycle {
    prevent_destroy = true
  }
}

output "admin_key" {
  value     = openai_admin_api_key.automation.value
  sensitive = true
}
```

## Expiration inputs

Configure at most one expiration input:

| Argument | Unit | Notes |
| --- | --- | --- |
| `expires_in_seconds` | Seconds | Sent directly to OpenAI. |
| `expire_in_hours` | Hours | Converted to seconds during create. |
| `expire_in_days` | Days | Converted to seconds during create. |

Omit all expiration inputs for a non-expiring key.

## Import and destroy behavior

| Resource | Import ID | Destroy behavior |
| --- | --- | --- |
| `openai_admin_api_key` | `admin_key_123` | Revokes/deletes the key through the OpenAI Admin API. |

Destroy and replacement require OpenAI to confirm `deleted=true` for the requested
admin key ID. If OpenAI does not confirm deletion, Terraform keeps the resource in
state and reports a revoke failure.

## Operational notes

- `value` is Sensitive and returned only during create, but it is still stored in Terraform state.
- Use encrypted remote state with restricted access for any workspace that creates keys.
- Add `lifecycle { prevent_destroy = true }` for critical or break-glass admin keys.
- Rotate by creating a new key, updating the external secret store/provider bootstrap path, then deliberately destroying the old key.
