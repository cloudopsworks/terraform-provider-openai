data "openai_organization_certificates" "all" {
  limit = 100
  order = "desc"
}
