resource "openai_project" "example" {
  name = "example-app"
}

resource "openai_service_account" "example" {
  project_id = openai_project.example.id
  name       = "example-app"
  role       = "member"
}
