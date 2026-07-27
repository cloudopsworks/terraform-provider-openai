resource "openai_organization_spend_alert" "eighty_percent" {
  threshold_amount = 200000 # USD cents

  notification_channel = {
    recipients     = ["platform-finops@example.com"]
    subject_prefix = "[OpenAI spend]"
  }
}
