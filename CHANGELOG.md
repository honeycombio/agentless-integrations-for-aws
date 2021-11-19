# agentless-integrations-for-aws Changelog

## 2.3.0 2021-11-19

### Features

- add debug mode for reading API responses (#80) | [@vreynolds](https://github.com/vreynolds)

### Maintenance

- docs: fix alb link (#85)
- docs: update key encryption (#81)
- remove vendor directory (#78)
- empower apply-labels action to apply labels (#74)
- Bump github.com/aws/aws-sdk-go from 1.40.34 to 1.41.7 (#72)
- Bump github.com/honeycombio/libhoney-go from 1.15.4 to 1.15.5 (#65)
- Bump github.com/aws/aws-lambda-go from 1.26.0 to 1.27.0 (#66)
- Bump github.com/aws/aws-sdk-go from 1.41.7 to 1.42.9 (#87)
- Bump github.com/honeycombio/honeytail from 1.0.0 to 1.6.0 (#77)

## 2.2.3 2021-10-18

### Fixes

- Revert "fix kms secret decryption" (#68) | [@JamieDanielson](https://github.com/JamieDanielson)

### Maintenance

- docs: add local testing docs (#69) | [@vreynolds](https://github.com/vreynolds)
- Change maintenance badge to maintained (#63) | [@JamieDanielson](https://github.com/JamieDanielson)
- Add Stalebot (#64) | [@JamieDanielson](https://github.com/JamieDanielson)
- maint: use latest go version in CI (#62) | [@vreynolds](https://github.com/vreynolds)
- Add NOTICE (#61) | [@cartermp](https://github.com/cartermp)
- Allow dependabot and forked PRs to run in CI (#56) | [@vreynolds](https://github.com/vreynolds)
- Add issue and PR templates (#55) | [@vreynolds](https://github.com/vreynolds)
- Add OSS lifecycle badge (#54) | [@vreynolds](https://github.com/vreynolds)
- Add community health files (#53) | [@vreynolds](https://github.com/vreynolds)
- correct the URL for the PR referenced (#50) | [@robbkidd](https://github.com/robbkidd)
- Bump github.com/aws/aws-sdk-go from 1.40.32 to 1.40.34 (#60)
- Bump github.com/sirupsen/logrus from 1.4.0 to 1.8.1 (#42)
- Bump github.com/aws/aws-sdk-go from 1.18.1 to 1.40.32 (#59)
- Bump github.com/honeycombio/libhoney-go from 1.9.3 to 1.15.4 (#58)
- Bump github.com/aws/aws-lambda-go from 1.9.0 to 1.26.0 (#52)

## 2.2.2 2021-07-08

- Update RDS PG log prefix to the only format allowed by RDS. [#37](https://github.com/honeycombio/agentless-integrations-for-aws/pull/37) | [@robbkidd](https://github.com/robbkidd)

## 2.2.1 2021-02-09

### Fixed

- Update to the S3 bucket log regex to match new field names and improve resilience to future changes [#33](https://github.com/honeycombio/agentless-integrations-for-aws/pull/33) | [@NFarrington](https://github.com/NFarrington)
- Fix KMS secret decryption [#34](https://github.com/honeycombio/agentless-integrations-for-aws/pull/34) | [@sbe-genomics](https://github.com/sbe-genomics)

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
