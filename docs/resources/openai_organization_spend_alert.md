# openai_organization_spend_alert

API group: [Organization Controls](../api-groups/organization-controls.md).

Creates and manages an OpenAI organization spend alert. Spend alerts notify email
recipients when organization spend reaches the configured threshold.

## Example Usage

```terraform
resource "openai_organization_spend_alert" "eighty_percent" {
  threshold_amount = 200000 # USD cents

  notification_channel = {
    recipients     = ["platform-finops@example.com"]
    subject_prefix = "[OpenAI spend]"
  }
}
```

## Import

```sh
terraform import openai_organization_spend_alert.eighty_percent spend_alert_123
```

## Schema

### Required

- `threshold_amount` (Number) Alert threshold amount in cents. OpenAI accepts zero or greater.
- `notification_channel` (Object) Email notification settings.
  - `recipients` (Set of String) Email addresses that receive alert notifications.

### Optional

- `currency` (String) Currency for `threshold_amount`. Default: `USD`.
- `interval` (String) Spend evaluation interval. Default: `month`.
- `notification_channel.type` (String) Notification channel type. Default: `email`.
- `notification_channel.subject_prefix` (String) Optional subject prefix for alert emails.

### Read-Only

- `id` (String) OpenAI spend alert ID.

## Notes

- Destroy deletes the spend alert through the OpenAI API.
- Recipient sets are sorted before state is written to keep plans stable.
