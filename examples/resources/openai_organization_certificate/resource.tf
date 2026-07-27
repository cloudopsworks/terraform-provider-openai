resource "openai_organization_certificate" "egress_root" {
  name        = "egress-root-ca"
  certificate = file("${path.module}/egress-root-ca.pem")
  active      = true
}
