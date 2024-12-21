packer {
  required_plugins {
    digitalocean = {
      source  = "github.com/hashicorp/digitalocean"
      version = "~> 1"
    }
    ansible = {
      version = "~> 1"
      source  = "github.com/hashicorp/ansible"
    }
  }
}

source "digitalocean" "app" {
  image        = "ubuntu-24-04-x64"
  region       = "nyc2"
  size         = "s-1vcpu-1gb"
  ssh_username = "root"
  temporary_key_pair_type = "ed25519"
}


build {
  sources = ["source.digitalocean.app"]

  provisioner "ansible" {
    playbook_file = "../ansible/packer_provision.yml"
    extra_arguments = [ "--scp-extra-args", "'-O'"] 
  }
}
