# Network switch definition
data "sakuracloud_switch" "your_switch" {
    filter {
        names = [ "YOUR_SWITCH_NAME" ]
    }
}
