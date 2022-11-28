# Honeycomb Agentless Integrations for AWS

[![OSS Lifecycle](https://img.shields.io/osslifecycle/honeycombio/agentless-integrations-for-aws?color=success)](https://github.com/honeycombio/home/blob/main/honeycomb-oss-lifecycle-and-practices.md)

This is a collection of AWS Lambda functions to collect log data from AWS services, transform it into structured events, and send it to [Honeycomb](https://honeycomb.io).

The easiest way to get started is to use either our [Terraform modules](https://github.com/honeycombio/terraform-aws-integrations) or [CloudFormation templates](https://github.com/honeycombio/cloudformation-integrations) to deploy these functions along with all the required AWS infrastructure.

## Features

### Logs in a Bucket

The `s3-handler` function will watch any S3 bucket for new files, parsing them into structured events and sending them to Honeycomb.
This function is setup with an S3 Bucket Notification triggering the function on each new file placed in the bucket.
It invokes `S3:GetObject` on the object, retrieves its contents, parses it line by line, and sends the resulting events directly to Honeycomb, using Honeycomb's custom event format.

Built-in parsers to transform unstructured log messages into structured events:
    - ELB access logs
    - ALB access logs
    - CloudFront access logs
    - S3 access logs
    - VPC flow logs
Additionally, a JSON parser can collect arbitrary [JSON lines](https://jsonlines.org) files and a generic regexp parser can structure custom log formats.

To get started, include the [s3-logfile Terraform module](https://registry.terraform.io/modules/honeycombio/integrations/aws/latest/submodules/s3-logfile) or launch the [s3-logfile CloudFormation stack](https://github.com/honeycombio/cloudformation-integrations#logs-from-a-s3-bucket).

### Kinesis Firehose Transforms

The `*-kfh-transform` functions act as Kinesis Firehose transform functions to take unstructured logs and parse them into structured events.
These functions act as stream processors, structuring Kinesis log records, which are ultimately be delivered to Honeycomb's [Kinesis Firehose HTTP endpoint](https://docs.honeycomb.io/getting-data-in/aws/how-aws-integrations-work/#how-aws-cloudwatch-logs-integrations-work).

Parsers exist for the following services:
- MySQL RDS slow query log
- PostgreSQL RDS logs

To get started, include the [rds-logs Terraform module](https://registry.terraform.io/modules/honeycombio/integrations/aws/latest/submodules/rds-logs) or launch the [rds-logs CloudFormation stack](https://github.com/honeycombio/cloudformation-integrations/blob/main/README.md#rds-cloudwatch-logs).
## Advanced Configuration

### Logs in a Bucket

### Environment Variables

The `s3-handler` accepts a number of of environment variables to configure behavior. Using Terraform or CloudFormation will set this values correctly.

Under most circumstances, custom setups will want to configure at minimum:

- `DATASET` - the name of the dataset to send events to.
- `HONEYCOMB_WRITE_KEY` (required) - your [Honeycomb API key](https://docs.honeycomb.io/getting-data-in/api-keys/) that has, at minimum, permissions to send events.
- `PARSER_TYPE` - built-in parsers like `alb` or `cloudfront` will configure for a known log format. `regex` will parse with an arbitrary user-defined regular expression. `keyval` will parse logs in `key=val` format. `json` will assume each line is a JSON object. See [here](https://github.com/honeycombio/agentless-integrations-for-aws/blob/main/common/common.go#L131-L157) for the full list of accepted values.

Advanced use cases may require these additional configuration options:

- `FILTER_FIELDS` - a comma-separated list of field names (JSON keys) to ignore, dropping them from any event that contains them. Useful to drop very large values or secret values.
- `REGEX_PATTERN` - with `PARSER_TYPE=regex`, will define the regular expression to use for parsing each line in the log file. Honeycomb columns are generated by defining named capture groups. For example, `(?P<name>[a-z]+)` would create a column called "name" if successfully parsed.
- `RENAME_FIELDS` - a comma-separate list of `before=after` pairs, where each `before` field will be renamed to `after`.
- `SAMPLE_RATE` - an integer > 0, indicating a sample rate. Only `1/SAMPLE_RATE` log messages will be kept, randomly.
- `TIME_FIELD_FORMAT` - a [Golang-compatible time formatting string](https://pkg.go.dev/time#Time.Format), specified as the reference time of `"01/02 03:04:05PM '06 -0700"`.
- `TIME_FIELD_NAME` - the name of the field to use as the timestamp for when this event occurred. If using a built-in parser like `alb` or `cloudfront`, will be set to an appropriate value.

#### Lambda Configuration

You should monitor your Lambda's execution metrics (hint: [Honeycomb can import CloudWatch Metrics](https://docs.honeycomb.io/getting-data-in/aws/how-aws-integrations-work/#metrics-via-aws-cloudwatch-metrics) to ensure it has sufficient memory to process logs and does not time out. Both the function's memory and timeout can be configured either in the AWS Console or via Terraform/CloudFormation.

## Debugging

First, check the function's logs (usually in CloudWatch Logs) to ensure no errors are reported.

If you don't see events in Honeycomb, there may be errors returned from the Honeycomb API. To see API responses, you can enable debugging
by adding the following environment variable to the lambda function created by the stack: `HONEYCOMB_DEBUG=true`

Something not working? Other questions? Create a GitHub Issue, or join our Slack community, Pollinators ([invite link](https://join.slack.com/t/honeycombpollinators/shared_invite/zt-xqexg936-dckd0l29wdE3WLmUs8Qvpg) to get help.

## Contributing

Features, bug fixes and other changes are gladly accepted.
Please open issues or a pull request with your change.
Remember to add your name to the CONTRIBUTORS file!

All contributions will be released under the Apache License 2.0.
