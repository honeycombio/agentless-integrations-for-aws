# This example should help you get going running the VPC flow log integration for Honeycomb
# It won't work out of the box - you will need to update some environment variables, and possibly tweak
# the configuration to work within your TF environment.
resource "aws_iam_role" "vpc_flow_log" {
  name = "vpc-flow-log-lambda-role"

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
  role   = "${aws_iam_role.vpc_flow_log.id}"
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
  role   = "${aws_iam_role.vpc_flow_log.id}"
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

resource "aws_lambda_function" "vpc_flow_log" {
  # change me to your region
  s3_bucket        = "honeycomb-integrations-us-east-1"
  s3_key           = "agentless-integrations-for-aws/LATEST/ingest-handlers.zip"
  function_name    = "honeycomb-vpc-flow-logs"
  role             = "${aws_iam_role.vpc_flow_log.arn}"
  handler          = "cloudwatch-handler"
  runtime          = "go1.x"
  memory_size      = "128"

  environment {
    variables = {
      ENVIRONMENT = "production" # change me
      PARSER_TYPE = "regex"
      # this pattern has been tested against the current version of VPC flow logs
      REGEX_PATTERN = "(?P<version>\\d+) (?P<account_id>\\d+) (?P<interface_id>eni-[0-9a-f]+) (?P<src_addr>[\\d\\.]+) (?P<dst_addr>[\\d\\.]+) (?P<src_port>\\d+) (?P<dst_port>\\d+) (?P<protocol>\\d+) (?P<packets>\\d+) (?P<bytes>\\d+) (?P<start_time>\\d+) (?P<end_time>\\d+) (?P<action>[A-Z]+) (?P<log_status>[A-Z]+)"
      # Change this to your encrypted Honeycomb write key or your raw write key (not recommended)
      HONEYCOMB_WRITE_KEY = "CHANGEME"
      # If the write key is encrypted, specify the KMS Key ID used to encrypt your write key
      # see https://github.com/honeycombio/agentless-integrations-for-aws#encrypting-your-write-key
      KMS_KEY_ID = "CHANGEME"
      DATASET = "vpc-flow-logs"
      SAMPLE_RATE = "100"
      TIME_FIELD_FORMAT = "%s(%L)?"
      TIME_FIELD_NAME = "start_time"
    }
  }
}

resource "aws_cloudwatch_log_subscription_filter" "flow_log_subscription_filter" {
  name            = "vpc-flow-log-subscription"
  log_group_name  = "/aws/vpc/flow-logs"
  filter_pattern  = ""
  destination_arn = "${aws_lambda_function.vpc_flow_log.arn}"
}

resource "aws_lambda_permission" "allow_cloudwatch" {
  statement_id   = "AllowExecutionFromCloudWatch"
  action         = "lambda:InvokeFunction"
  function_name  = "${aws_lambda_function.vpc_flow_log.arn}"
  principal      = "logs.amazonaws.com"
}

## Uncomment these sections if you haven't configured VPC flow logs
# resource "aws_flow_log" "flow_log" {
#   log_group_name = "${aws_cloudwatch_log_group.log_group.name}"
#   iam_role_arn   = "${aws_iam_role.flow_log_role.arn}"
#   # change this to your VPC ID
#   vpc_id         = "${aws_vpc.id}"
#   traffic_type   = "ALL"
# }

# resource "aws_cloudwatch_log_group" "log_group" {
#   name = "/aws/vpc/flow-logs"
# }

# resource "aws_iam_role" "flow_log_role" {
#   name = "flow-log-role"

#   assume_role_policy = <<EOF
# {
#   "Version": "2012-10-17",
#   "Statement": [
#     {
#       "Sid": "",
#       "Effect": "Allow",
#       "Principal": {
#         "Service": "vpc-flow-logs.amazonaws.com"
#       },
#       "Action": "sts:AssumeRole"
#     }
#   ]
# }
# EOF
# }

# # necessary for vpc flow logs to write to cloudwatch
# resource "aws_iam_role_policy" "flow_log_policy" {
#   name = "flow-log-policy"
#   role = "${aws_iam_role.flow_log_role.id}"

#   policy = <<EOF
# {
#   "Version": "2012-10-17",
#   "Statement": [
#     {
#       "Action": [
#         "logs:CreateLogGroup",
#         "logs:CreateLogStream",
#         "logs:PutLogEvents",
#         "logs:DescribeLogGroups",
#         "logs:DescribeLogStreams"
#       ],
#       "Effect": "Allow",
#       "Resource": "*"
#     }
#   ]
# }
# EOF
# }
