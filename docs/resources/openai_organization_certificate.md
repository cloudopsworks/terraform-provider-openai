# openai_organization_certificate

API group: [Organization Controls](../api-groups/organization-controls.md).

Uploads an OpenAI organization certificate and optionally activates it at the
organization scope. OpenAI does not auto-activate uploaded certificates.

## Example Usage

```terraform
resource "openai_organization_certificate" "egress_root" {
  name        = "egress-root-ca"
  certificate = file("${path.module}/egress-root-ca.pem")
  active      = true
}
```

## Import

```sh
terraform import openai_organization_certificate.egress_root cert_123
```

After import, set `active` to the desired state. `certificate` is not returned by
normal refresh and is only needed when Terraform creates/replaces the certificate.

## Schema

### Optional

- `name` (String) Certificate name. OpenAI supports updating only the name after upload.
- `certificate` (String, Sensitive) Certificate content in PEM format. Required for create; changes replace the certificate.
- `active` (Boolean) Whether the certificate should be active at the organization scope. Default: `false`.

### Read-Only

- `id` (String) OpenAI certificate ID.
- `object` (String) OpenAI certificate object type.
- `created_at` (Number) Unix timestamp when the certificate was uploaded.
- `certificate_details` (Object) Validity metadata returned by OpenAI.
  - `content` (String, Sensitive) Usually null for the resource; use the single-certificate data source with `include_content = true` if PEM retrieval is required.
  - `expires_at` (Number) Unix timestamp when the certificate expires, when returned.
  - `valid_at` (Number) Unix timestamp when the certificate becomes valid, when returned.

## Notes

- Destroy deactivates the certificate first when `active = true`, then deletes it.
- List data sources omit PEM content. Avoid exposing certificate contents in outputs.
