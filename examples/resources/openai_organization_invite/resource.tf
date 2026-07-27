resource "openai_project" "platform" {
  name = "platform"
}

resource "openai_organization_invite" "platform_admin" {
  email = "platform-admin@example.com"
  role  = "owner"

  projects = [{
    id   = openai_project.platform.id
    role = "owner"
  }]
}
