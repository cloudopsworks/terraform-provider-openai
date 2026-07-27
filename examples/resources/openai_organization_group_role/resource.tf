resource "openai_organization_group_role" "engineering_auditor" {
  group_id = openai_organization_group.engineering.id
  role_id  = openai_organization_role.auditor.id
}
