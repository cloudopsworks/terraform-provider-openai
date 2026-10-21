resource "openai_project_user" "operator" {
  project_id = var.openai_project_id
  user_id    = var.openai_user_id
  role       = "member"
}
