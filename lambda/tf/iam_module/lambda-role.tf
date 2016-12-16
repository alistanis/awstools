resource "aws_iam_role" "lambda_execute_role" {
  name = "lambda_execute_role"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "lambda_ec2_attachment" {
  role = "${aws_iam_role.lambda_execute_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2FullAccess"
}

resource "aws_iam_role_policy_attachment" "lambda_s3_attachment" {
  role = "${aws_iam_role.lambda_execute_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/AmazonS3FullAccess"
}

resource "aws_iam_role_policy_attachment" "lambda_lambda_attachment" {
  role = "${aws_iam_role.lambda_execute_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/AWSLambdaFullAccess"
}

resource "aws_iam_role_policy_attachment" "lambda_cloudwatch_attachment" {
  role = "${aws_iam_role.lambda_execute_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchFullAccess"
}


output "lambda_execute_arn" {value = "${aws_iam_role.lambda_execute_role.arn}"}