# openai_organization_data_retention

API group: [Organization Controls](../api-groups/organization-controls.md).

OpenAI API hierarchy: `Administration > Organization > Data Retention`.

Reads the current OpenAI organization data-retention mode.

## Example Usage

```terraform
data "openai_organization_data_retention" "current" {}
```

## Schema

### Read-Only

- `id` (String) Synthetic singleton ID, always `organization`.
- `type` (String) Organization data-retention mode.
