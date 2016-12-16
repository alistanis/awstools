module "s3" {
  source = "../s3"
  #name = "s3"
  bucket_name = "${var.bucket_name}"
}

module "scheduled_lambda_module" {
  source = "github.com/alistanis/awstools//lambda/tf/scheduled_lambda_module"
 # name = "create_lambda_function"
  lambda_function_suffix = "snapshots"
  lambda_s3_bucket = "${module.s3.s3_bucket_name}"
  rate_name = "once-daily"
  rate_description = "This event will fire once daily at the specified time."
  rate_schedule_expression = "cron(* 20 * * ? *)"
  cloudwatch_event_target_id = "LambdaScheduledSnapshots"
}