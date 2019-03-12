resource "aws_iam_role" "honeycomb_mysql_rds_logs_role" {
  name = "honeycomb-mysql-rds-logs-role"

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

resource "aws_iam_role_policy" "mysql_rds_lambda_log_policy" {
  name   = "honeycomb-mysql-rds-lambda-logs-policy"
  role   = "${aws_iam_role.honeycomb_mysql_rds_logs_role.id}"
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

resource "aws_iam_role_policy" "mysql_rds_lambda_kms_policy" {
  name   = "honeycomb-mysql-rds-lambda-kms-policy"
  role   = "${aws_iam_role.honeycomb_mysql_rds_logs_role.id}"
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

resource "aws_lambda_function" "honeycomb_mysql_rds_logs" {
  # change me to your region
  s3_bucket        = "honeycomb-integrations-us-east-1"
  s3_key           = "agentless-integrations-for-aws/LATEST/ingest-handlers.zip"
  function_name    = "honeycomb-mysql-rds-logs"
  role             = "${aws_iam_role.honeycomb_mysql_rds_logs_role.arn}"
  handler          = "mysql-handler"
  runtime          = "go1.x"
  memory_size      = "128"

  environment {
    variables = {
      ENVIRONMENT = "production" # change me
      # Change this to your encrypted Honeycomb write key or your raw write key (not recommended)
      HONEYCOMB_WRITE_KEY = "CHANGEME"
      # If the write key is encrypted, specify the KMS Key ID used to encrypt your write key
      # see https://github.com/honeycombio/agentless-integrations-for-aws#encrypting-your-write-key
      KMS_KEY_ID = "CHANGEME"
      DATASET = "mysql-rds-logs"
      SAMPLE_RATE = "1"
      SCRUB_QUERY = "false"
    }
  }
}

resource "aws_cloudwatch_log_subscription_filter" "honeycomb_mysql_rds_lambdafunction_logfilter" {
  name            = "honeycomb-mysql-rds-subscription"
  # set this to the log group name associated with your instance
  log_group_name  = "/aws/rds/instance/CHANGEME/slowquery"
  filter_pattern  = ""
  destination_arn = "${aws_lambda_function.honeycomb_mysql_rds_logs.arn}"
}

resource "aws_lambda_permission" "allow_cloudwatch" {
  statement_id   = "AllowExecutionFromCloudWatch"
  action         = "lambda:InvokeFunction"
  function_name  = "${aws_lambda_function.honeycomb_mysql_rds_logs.arn}"
  principal      = "logs.amazonaws.com"
}
