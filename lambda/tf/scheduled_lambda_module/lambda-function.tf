//----------
// Variables
//----------

# lambda function variables
variable "lambda_function_suffix" {}
variable "lambda_s3_bucket" {}

variable "lambda_handler" {
  default = "handler.handle"
}

# lambda scheduling variables
variable "rate_name" {}
variable "rate_description" {}
variable "rate_schedule_expression" {}
variable "cloudwatch_event_target_id" {}

variable "runtime" {
  default = "python2.7"
}

variable "timeout" {
  default = 300
}

//----------------
// Call IAM Module
//----------------
module "iam_module" {
  source = "../iam_module"
}

//--------------------------
// Lambda specific resources
//--------------------------

resource "aws_lambda_function" "function" {
  s3_bucket = "${var.lambda_s3_bucket}"
  function_name = "${join("", list("lambda-scheduled_", var.lambda_function_suffix))}"
  role = "${module.iam_module.lambda_execute_arn}"
  handler = "${var.lambda_handler}"
  runtime = "${var.runtime}"
  timeout = "${var.timeout}"
}

resource "aws_cloudwatch_event_rule" "rate" {
  name = "${var.rate_name}"
  description = "${var.rate_description}"
  schedule_expression = "${var.rate_schedule_expression}"
}

resource "aws_cloudwatch_event_target" "run_lambda_on_schedule" {
  rule = "${aws_cloudwatch_event_rule.rate.name}"
  target_id = "${var.cloudwatch_event_target_id}"
  arn = "${aws_lambda_function.function.arn}"
}

resource "aws_lambda_permission" "allow_cloudwatch_to_call_lambda_function" {
  statement_id = "AllowExecutionFromCloudWatch"
  action = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.function.function_name}"
  principal = "events.amazonaws.com"
  source_arn = "${aws_cloudwatch_event_rule.rate.arn}"
}