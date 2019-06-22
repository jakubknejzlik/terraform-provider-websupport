provider "websupport" {
  username = var.websupport_username
  password = var.websupport_password
}

resource "websupport_record" "terraform_dns_record" {
  zone  = "example.com"
  name  = "terraform"
  value = "192.168.0.22"
  type  = "A"
  ttl   = "3600"
}