// region variable
variable "region" {
  description = "The region where the resources will be created"
  type        = string
  default     = "nyc2"
}

data "digitalocean_image" "base_image" {
  name = "packer-1728353739"
}

resource "digitalocean_droplet" "app" {
  image  = data.digitalocean_image.base_image.image
  name   = "web-1"
  region = var.region
  size   = "s-1vcpu-1gb"
  ssh_keys = [
    data.digitalocean_ssh_key.terraform.id,
  ]
  connection {
    host        = self.ipv4_address
    user        = "root"
    type        = "ssh"
    private_key = file(var.pvt_key)
    timeout     = "2m"
  }
}

data "digitalocean_ssh_key" "terraform" {
  name = "do"
}

resource "digitalocean_domain" "default" {
  name       = "movies-with-friends.com"
  ip_address = digitalocean_droplet.app.ipv4_address
}

resource "digitalocean_volume" "app" {
  region                  = var.region
  name                    = "moviesdb"
  size                    = 25
  initial_filesystem_type = "ext4"
  description             = "volume for movies db"
}


resource "digitalocean_volume_attachment" "app" {
  droplet_id = digitalocean_droplet.app.id
  volume_id  = digitalocean_volume.app.id
  depends_on = [
    digitalocean_volume.app,
    digitalocean_droplet.app
  ]
}
