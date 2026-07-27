resource "openai_organization_spend_limit" "monthly" {
  threshold_amount = 250000 # USD cents
}
