resource "openai_admin_api_key" "automation" {
  name            = "terraform-automation"
  expire_in_hours = 10
  lifecycle {
    prevent_destroy = true
  }
}
