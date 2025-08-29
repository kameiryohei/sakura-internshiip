# 認証情報
variable "sakuracloud_token" {
    type        = string
    description = "さくらのクラウドAPIトークン"
}

variable "sakuracloud_secret" {
    type        = string
    description = "さくらのクラウドAPIシークレット"
}

# リソース設定
variable "prefix" {
    type        = string
    default     = "github-actions-tf"
    description = "リソース名のプレフィックス"
}

variable "image_id" {
    type        = string
    description = "Ubuntu Server 24.04 Cloud ImageのアーカイブID"
    default     = "112900898555"
}

variable "ssh_public_key" {
    type        = string
    description = "VMに設定するSSH公開鍵"
}

variable "zone" {
    type        = string
    description = "デプロイするゾーン"
    default     = "is1a"
}