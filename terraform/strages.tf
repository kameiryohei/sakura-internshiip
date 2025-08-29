# VM disk definition
resource "sakuracloud_disk" "ubuntu_disk" {
    name              = "${var.prefix}-disk"
    plan              = "20"
    source_archive_id = var.image_id
}