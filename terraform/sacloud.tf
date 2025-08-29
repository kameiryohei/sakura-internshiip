# Terraform provider configuration
terraform {
    required_providers {
        sakuracloud = {
            source  = "sacloud/sakuracloud"
            version = "~> 2.18"
        }
    }
}

provider "sakuracloud" {
    token  = var.sakuracloud_token
    secret = var.sakuracloud_secret
    zone   = var.zone
}


# VM instance definition
resource "sakuracloud_server" "nethygiene" {
    name   = "${var.prefix}-server"
    plan   = "2core-4gb"
    core   = 2
    memory = 4
    disks  = [sakuracloud_disk.ubuntu_disk.id]
    
    user_data = templatefile("${path.module}/nethygiene.yaml", {
        SSH_PUBLIC_KEY = var.ssh_public_key
    })
    
    network_interface {
        upstream = data.sakuracloud_switch.your_switch.id
    }
    
    boot_after_create = true
}
