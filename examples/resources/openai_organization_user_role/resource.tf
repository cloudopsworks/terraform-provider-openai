resource "openai_organization_user_role" "alice_auditor" {
  user_id = data.openai_organization_user.alice.id
  role_id = openai_organization_role.auditor.id
}
