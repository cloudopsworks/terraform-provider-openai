resource "openai_organization_role" "auditor" {
  name        = "auditor"
  description = "Read-only organization auditor"
  permissions = [
    "organization.users.read",
  ]
}
