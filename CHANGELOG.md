# agentless-integrations-for-aws Changelog

## 2.2.0 2020-06-04

- Mysql and Postgres integration now include cloudwatch metdata [#32](https://github.com/honeycombio/agentless-integrations-for-aws/pull/32)

## 2.0.2 2019-11-15

- Fixes an issue [with panics on invalid JSON
  payloads](https://github.com/honeycombio/agentless-integrations-for-aws/pull/25).

## 2.0.1 2019-08-18

- Updates ALB regex template to support newer ALB fields. See [#24](https://github.com/honeycombio/agentless-integrations-for-aws/pull/24).

## 2.0.0 2019-09-04

Major version increase due to a breaking change in the S3 Integration. In the past we've tried to support processing gzip-compressed objects based on its `Content-Type` header or filename extension. This was dubious because these aren't reliable indicators of whether or not content is gzipped. The correct way to store gzipped content is to set the `Content-Encoding` header in S3 with your object. The S3 handler now looks at this header exclusively. If you have gzipped content in your buckets that is not stored with the correct `Content-Encoding`, you can still force processing using gzip by setting the new `ForceGunzip` parameter to `true` in the Cloudformation template, or by setting `FORCE_GUNZIP=true` in the lambda environment. The handler will attempt to decompress the object, then fall back to uncompressed processing if that fails.

## 1.9.0

Features

- New `FilterFields` option in Cloudwatch and S3 logs handler Cloudformation templates (environment variable is `FILTER_FIELDS` in the Lambda environment). When set to a comma-separated list of strings, will drop fields that match any field name present in the list. You can use this to prevent sensitive data from being sent to Honeycomb. For example, if your log files have the fields named "address" and "zip_code" and you want to drop them, pass a `FilterFields` value of `address,zip_code`.

## 1.8.0

Features

- Allows override of the scan buffer size in the S3 handler. If your log files contain very large lines (> 64KiB), you can set `BUFFER_SIZE` in the lambda environment to a larger value. The S3 Logs Cloudformation Template now also accepts a `BufferSize` parameter.

## 1.7.0

Features

- The AWS Publisher now includes a `aws.cloudwatch.logstream` field indicating the Cloudwatch Log Stream name that was the source of incoming events.
