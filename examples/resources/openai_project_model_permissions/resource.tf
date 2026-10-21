resource "openai_project_model_permissions" "production" {
  project_id = var.openai_project_id
  mode       = "allow_list"
  model_ids  = ["gpt-4.1", "gpt-4o"]
}
