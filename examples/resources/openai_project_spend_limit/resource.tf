resource "openai_project_spend_limit" "monthly" {
  project_id       = var.openai_project_id
  threshold_amount = 250000 # USD cents
}
