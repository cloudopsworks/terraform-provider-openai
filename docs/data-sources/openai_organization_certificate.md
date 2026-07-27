# openai_organization_certificate

API group: [Organization Controls](../api-groups/organization-controls.md).

OpenAI API hierarchy: `Administration > Organization > Certificates`.

Reads one OpenAI organization certificate by ID. Set `include_content = true` only
when the PEM content is needed and the Terraform state backend is protected.

## Example Usage

```terraform
data "openai_organization_certificate" "egress_root" {
  id              = "cert_123"
  include_content = false
}
```

## Schema

### Required

- `id` (String) OpenAI certificate ID.

### Optional

- `include_content` (Boolean) Request `certificate_details.content` from OpenAI and store it as Sensitive state.

### Read-Only

- `name` (String) Certificate name.
- `object` (String) OpenAI certificate object type.
- `active` (Boolean) Whether the certificate is active at organization scope.
- `created_at` (Number) Unix timestamp when the certificate was uploaded.
- `certificate_details` (Object) Certificate validity metadata.
  - `content` (String, Sensitive) Certificate content when `include_content = true` and OpenAI returns it.
  - `expires_at` (Number) Unix timestamp when the certificate expires, when returned.
  - `valid_at` (Number) Unix timestamp when the certificate becomes valid, when returned.
