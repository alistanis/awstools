variable "bucket_name" {}
variable "bucket_permissions" {
  default = "private"
}

variable "lambda_payload_file" { }
variable "s3_key" {
  default = "handler.zip"
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


// This could be dangerous - make sure you test your code before using this line because it'll deploy the lambda function
// that being said, versioning is turned on in the s3 bucket
resource "aws_s3_bucket_object" "lambda_payload" {
  bucket = "${var.bucket_name}"
  key = "${var.s3_key}"
  source = "${var.lambda_payload_file}"
  etag = "${md5(file(var.lambda_payload_file))}"
}

output "s3_bucket_name" {value= "${aws_s3_bucket.s3_bucket.bucket}"}