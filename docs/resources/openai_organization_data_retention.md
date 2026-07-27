# openai_organization_data_retention

API group: [Organization Controls](../api-groups/organization-controls.md).

Manages the organization-wide OpenAI data-retention mode. This is a singleton
setting and uses the synthetic Terraform ID `organization`.

## Example Usage

```terraform
resource "openai_organization_data_retention" "default" {
  type = "modified_abuse_monitoring"
}
```

## Import

```sh
terraform import openai_organization_data_retention.default organization
```

## Schema

### Required

- `type` (String) Organization data-retention mode. Supported values: `zero_data_retention`, `modified_abuse_monitoring`, `enhanced_zero_data_retention`, `enhanced_modified_abuse_monitoring`.

### Read-Only

- `id` (String) Synthetic singleton ID, always `organization`.

## Notes

- Create and update both call OpenAI's data-retention update endpoint.
- Destroy removes only the Terraform state object and emits a warning because OpenAI exposes retrieve/update but no delete/reset endpoint for organization data retention.
