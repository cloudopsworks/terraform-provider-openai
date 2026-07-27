data "openai_admin_api_keys" "all" {
  limit = 100
  order = "desc"
}
