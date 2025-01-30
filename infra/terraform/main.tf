// region variable
variable "region" {
  description = "The region where the resources will be created"
  type        = string
  default     = "nyc2"
}

data "digitalocean_image" "base_image" {
  name = "packer-1730661392"
}

resource "digitalocean_droplet" "app" {
  image  = data.digitalocean_image.base_image.image
  name   = "web-1"
  region = var.region
  size   = "s-1vcpu-1gb"
  ssh_keys = [
    data.digitalocean_ssh_key.terraform.id,
  ]


  # Wait for cloud-init to complete
  provisioner "remote-exec" {
    inline = ["echo 'Waiting for cloud-init to complete...'"]

    connection {
      type        = "ssh"
      host        = self.ipv4_address
      user        = "root"
      private_key = file("~/.ssh/do")
      timeout     = "2m"
    }
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

# Create a new container registry
resource "digitalocean_container_registry" "registry" {
  name                   = "jmaguireregistry"
  subscription_tier_slug = "starter"
  region                 = "nyc2"
}

# Create a container registry docker credentials
resource "digitalocean_container_registry_docker_credentials" "registry_creds" {
  registry_name = digitalocean_container_registry.registry.name
  write         = true # Set to true for read/write access, false for read-only
}

# outputs.tf
output "vps_ip" {
  value = digitalocean_droplet.app.ipv4_address
}

output "registry_endpoint" {
  value = digitalocean_container_registry.registry.endpoint
}

output "registry_server_url" {
  value = digitalocean_container_registry.registry.server_url
}

output "registry_credentials" {
  value     = digitalocean_container_registry_docker_credentials.registry_creds.docker_credentials
  sensitive = true # Marks the output as sensitive to prevent credentials from showing in logs
}
