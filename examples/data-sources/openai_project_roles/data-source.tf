data "openai_project_roles" "app" {
  project_id = "proj_123"
  limit      = 100
}
