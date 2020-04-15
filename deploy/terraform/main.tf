variable "hcloud_token" {
  type = string
}

variable "gandi_key" {
  type = string
}

provider "hcloud" {
  token = var.hcloud_token
}

provider "gandi" {
  key = var.gandi_key
}

provider "acme" {
  server_url = "https://acme-staging-v02.api.letsencrypt.org/directory"
}

data "template_file" "cloudinit" {
  template = file("${path.module}/cloud-config.yml")

  vars = {
    ip_address = hcloud_floating_ip.master.ip_address
    tls_cert = acme_certificate.certificate.certificate_pem
    tls_key = acme_certificate.certificate.private_key_pem
  }
}

data "hcloud_ssh_keys" "all_keys" {
}

resource "hcloud_server" "web" {
  name = "apparea"
  server_type = "cx11"
  image = "ubuntu-18.04"
  location = "nbg1"

  ssh_keys = data.hcloud_ssh_keys.all_keys.ssh_keys.*.name

  user_data = data.template_file.cloudinit.rendered
}

resource "hcloud_floating_ip" "master" {
  type = "ipv4"
  home_location = "nbg1"
}

resource "hcloud_floating_ip_assignment" "master" {
  floating_ip_id = hcloud_floating_ip.master.id
  server_id = hcloud_server.web.id
}

data "gandi_livedns_domain" "apparea_dev" {
  name = "apparea.dev"
}

resource "gandi_livedns_record" "apparea_record" {
  zone = data.gandi_livedns_domain.apparea_dev.name
  name = "@"
  type = "A"
  ttl = 3600
  values = [
    hcloud_floating_ip.master.ip_address
  ]
}

resource "tls_private_key" "private_key" {
  algorithm = "RSA"
}

resource "acme_registration" "reg" {
  account_key_pem = tls_private_key.private_key.private_key_pem
  email_address   = "jedevc@gmail.com"
}

resource "acme_certificate" "certificate" {
  account_key_pem           = acme_registration.reg.account_key_pem
  common_name               = "*.apparea.dev"

  dns_challenge {
    provider = "gandiv5"

    config = {
      GANDIV5_API_KEY = var.gandi_key
    }
  }
}
