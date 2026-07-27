data "openai_organization_spend_alerts" "all" {
  limit = 100
  order = "desc"
}
