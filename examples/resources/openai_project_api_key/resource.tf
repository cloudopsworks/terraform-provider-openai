resource "openai_project" "example" {
  name = "example-app"
}

resource "openai_service_account" "example" {
  project_id = openai_project.example.id
  name       = "example-app"
  role       = "member"
}

resource "openai_project_api_key" "example" {
  project_id         = openai_project.example.id
  service_account_id = openai_service_account.example.id
  name               = "example-app"
}

output "project_api_key" {
  value     = openai_project_api_key.example.value
  sensitive = true
}
