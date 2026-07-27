# openai_organization_spend_limit

API group: [Organization Controls](../api-groups/organization-controls.md).

Manages the organization hard monthly spend limit. This is a singleton setting
and uses the synthetic Terraform ID `organization`.

## Example Usage

```terraform
resource "openai_organization_spend_limit" "monthly" {
  threshold_amount = 250000 # USD cents
}
```

## Import

```sh
terraform import openai_organization_spend_limit.monthly organization
```

## Schema

### Required

- `threshold_amount` (Number) Hard spend limit amount in cents.

### Optional

- `currency` (String) Currency for `threshold_amount`. Default: `USD`.
- `interval` (String) Spend evaluation interval. Default: `month`.

### Read-Only

- `id` (String) Synthetic singleton ID, always `organization`.
- `enforcement_status` (String) OpenAI enforcement status for the hard spend limit, such as `inactive` or `enforcing`.

## Notes

- Create and update both call OpenAI's spend-limit update endpoint.
- Destroy deletes the organization spend limit through the OpenAI API.
