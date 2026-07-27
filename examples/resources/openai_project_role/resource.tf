resource "openai_project_role" "api_key_reader" {
  project_id  = openai_project.example.id
  name        = "api-key-reader"
  description = "Read project API keys"
  permissions = [
    "project.api_keys.read",
  ]
}
