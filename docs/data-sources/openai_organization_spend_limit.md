# openai_organization_spend_limit

API group: [Organization Controls](../api-groups/organization-controls.md).

Reads the current OpenAI organization hard spend limit.

## Example Usage

```terraform
data "openai_organization_spend_limit" "current" {}
```

## Schema

### Read-Only

- `id` (String) Synthetic singleton ID, always `organization`.
- `threshold_amount` (Number) Hard spend limit amount in cents.
- `currency` (String) Currency for `threshold_amount`.
- `interval` (String) Spend evaluation interval.
- `enforcement_status` (String) OpenAI enforcement status for the hard spend limit.
