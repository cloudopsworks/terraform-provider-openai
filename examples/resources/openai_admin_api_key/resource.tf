resource "openai_admin_api_key" "automation" {
  name = "terraform-automation"

  lifecycle {
    prevent_destroy = true
  }
}
