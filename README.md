Terraform Provider
==================

Build from root of this project:
```
$ go build -o ./example/terraform-provider-websupport
```

Edit main.tf and variables.tf
```
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
```

Apply
```
$ cd example
$ terraform init
$ terraform plan
$ terraform apply
```