data "openai_service_account" "ops" {
  project_id = "proj_abc123"
  name       = "ops"
}

data "openai_service_account" "by_id" {
  project_id = "proj_abc123"
  id         = "svc_acct_abc123"
}
