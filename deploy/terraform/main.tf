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

data "template_file" "cloudinit" {
  template = file("${path.module}/cloud-config.yml")

  vars = {
    ip_address = hcloud_floating_ip.master.ip_address
    gandi_api_key = var.gandi_key
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

resource "gandi_livedns_record" "apparea_recursive_record" {
  zone = data.gandi_livedns_domain.apparea_dev.name
  name = "*"
  type = "A"
  ttl = 3600
  values = [
    hcloud_floating_ip.master.ip_address
  ]
}
