resource "openai_organization_group_user" "engineering_alice" {
  group_id = openai_organization_group.engineering.id
  user_id  = data.openai_organization_user.alice.id
}
