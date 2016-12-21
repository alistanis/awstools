// set region
provider "aws" {
  region = "us-east-1"
}

// call s3 module to set up buckets/keys
module "s3" {
  source = "../s3"
  # use a different bucket
  bucket_name = "hcdo-scheduled-lambda-ami-snapshots"
  lambda_payload_file = "../../ami-snapshots/functions/create_snapshot/handler.zip"
}

// call the lambda module which will call the iam module
module "scheduled_lambda_module" {
  source = "../scheduled_lambda_module"
 # name = "create_lambda_function"
  lambda_function_suffix = "snapshots"
  lambda_s3_bucket = "${module.s3.s3_bucket_name}"
  rate_name = "once-daily"
  rate_description = "This event will fire once daily at the specified time."
  // this will launch once a day at 9PM
  rate_schedule_expression = "cron(* 20 * * ? *)"
  cloudwatch_event_target_id = "LambdaScheduledSnapshots"
}