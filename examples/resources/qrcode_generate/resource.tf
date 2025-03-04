# Copyright (c) HashiCorp, Inc.

resource "qrcode_generate" "default" {
  file = "/tmp/qrcode.png"
  text = "qrcode"
}
