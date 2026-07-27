# openai_organization_spend_alert

API group: [Organization Controls](../api-groups/organization-controls.md).

OpenAI API hierarchy: `Administration > Organization > Spend Alerts`.

Reads one OpenAI organization spend alert by ID.

## Example Usage

```terraform
data "openai_organization_spend_alert" "eighty_percent" {
  id = "spend_alert_123"
}
```

## Schema

### Required

- `id` (String) OpenAI spend alert ID.

### Read-Only

- `threshold_amount` (Number) Alert threshold amount in cents.
- `currency` (String) Currency for `threshold_amount`.
- `interval` (String) Spend evaluation interval.
- `notification_channel` (Object) Email notification settings.
  - `recipients` (Set of String) Email addresses that receive alert notifications.
  - `type` (String) Notification channel type.
  - `subject_prefix` (String) Optional subject prefix.
