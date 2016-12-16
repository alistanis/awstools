variable "bucket_name" {}
variable "bucket_permissions" {
  default = "private"
}

variable "versioning" {
  default = true
}

resource "aws_s3_bucket" "s3_bucket" {
  bucket = "${var.bucket_name}"
  acl = "${var.bucket_permissions}"
  versioning {
    enabled = "${var.versioning}"
  }
}

output "s3_bucket_name" {value= "${aws_s3_bucket.s3_bucket.bucket}"}