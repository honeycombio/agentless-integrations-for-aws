# This example should help you get going running the generic JSON integration for Cloudwatch ogs
# It won't work out of the box - you will need to update some environment variables, and possibly tweak
# the configuration to work within your TF environment.
resource "aws_iam_role" "honeycomb_cloudwatch_logs" {
  name = "honeycomb-cloudwatch-logs-lambda-role"

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
resource "aws_iam_role_policy" "lambda_log_policy" {
  name   = "lambda-logs-policy"
  role   = "${aws_iam_role.honeycomb_cloudwatch_logs.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Effect": "Allow",
      "Resource": "arn:aws:logs:*:*:*"
    }
  ]
}
EOF
}

# This section can be omitted if you opt to not use an encrypted write key.
# Otherwise, change the policy below to use the KMS Key ID used to encrypt your
# write key. See https://github.com/honeycombio/agentless-integrations-for-aws#encrypting-your-write-key
resource "aws_iam_role_policy" "lambda_kms_policy" {
  name   = "lambda-kms-policy"
  role   = "${aws_iam_role.honeycomb_cloudwatch_logs.id}"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "kms:Decrypt",
      "Effect": "Allow",
      "Resource": "arn:aws:kms:*:*:key/CHANGEME"
    }
  ]
}
EOF
}

resource "aws_lambda_function" "cloudwatch_logs" {
  # change me to your region
  s3_bucket        = "honeycomb-integrations-us-east-1"
  s3_key           = "agentless-integrations-for-aws/LATEST/ingest-handlers.zip"
  function_name    = "honeycomb-cloudwatch-logs-integration"
  role             = "${aws_iam_role.honeycomb_cloudwatch_logs.arn}"
  handler          = "cloudwatch-handler"
  runtime          = "go1.x"
  memory_size      = "128"

  environment {
    variables = {
      ENVIRONMENT = "production" # change me
      PARSER_TYPE = "json"
      # Change this to your encrypted Honeycomb write key or your raw write key (not recommended)
      HONEYCOMB_WRITE_KEY = "CHANGEME"
      # If the write key is encrypted, specify the KMS Key ID used to encrypt your write key
      # see https://github.com/honeycombio/agentless-integrations-for-aws#encrypting-your-write-key
      KMS_KEY_ID = "CHANGEME"
      DATASET = "cloudwatch-logs"
      SAMPLE_RATE = "100"
      TIME_FIELD_FORMAT = "%s(%L)?"
      TIME_FIELD_NAME = "start_time"
    }
  }
}

resource "aws_cloudwatch_log_subscription_filter" "cloudwatch_subscription_filter" {
  name            = "log-group-subscription"
  # change this to the target log group
  log_group_name  = "/target/log/group"
  filter_pattern  = ""
  destination_arn = "${aws_lambda_function.cloudwatch_logs.arn}"
}

resource "aws_lambda_permission" "allow_cloudwatch" {
  statement_id   = "AllowExecutionFromCloudWatch"
  action         = "lambda:InvokeFunction"
  function_name  = "${aws_lambda_function.cloudwatch_logs.arn}"
  principal      = "logs.amazonaws.com"
}
