resource "openai_project" "example" {
  name = "example-app"
}

resource "openai_service_account" "example" {
  project_id = openai_project.example.id
  name       = "example-app"
  role       = "member"
  scopes     = ["responses.read"]
}

output "service_account_api_key" {
  value     = openai_service_account.example.api_key_value
  sensitive = true
}
