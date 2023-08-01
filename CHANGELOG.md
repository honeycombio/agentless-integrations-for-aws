# agentless-integrations-for-aws Changelog

## 3.2.0 2023-08-01

### Fixed

- S3 Handler: Handle all variations of AWS X-Ray trace id and error on unparsed events (#214) | [@brookesargent](https://github.com/brookesargent)

### Maintenance

- Bump github.com/aws/aws-sdk-go from 1.44.273 to 1.44.294 (#211)
- Bump github.com/honeycombio/libhoney-go from 1.18.0 to 1.20.0 (#212)
- Bump github.com/sirupsen/logrus from 1.9.2 to 1.9.3 (#213)
- Bump github.com/aws/aws-sdk-go from 1.44.294 to 1.44.313 (#215)

## 3.1.0 2023-06-08

### Enhancement

- Adds support for a SAMPLE_RATE_RULES environment variable for the s3 handler, 
  allowing you to apply different sample rates for different bucket prefixes. (#205) | [@NLincoln](https://github.com/NLincoln)

## 3.0.1 2022-12-02

### Fixed

- KFH RDS Transform: pass non-CloudWatch Log data through unaltered (#178) | [@jharley](https://github.com/jharley)
- KFH RDS Transform: mark unparsable records as failures (#179) | [@jharley](https://github.com/jharley)

### Maintenance
- Bump github.com/honeycombio/honeytail from 1.8.1 to 1.8.2 (#174) | [@dependabot](https://github.com/dependabot)
- Bump github.com/aws/aws-lambda-go from 1.34.1 to 1.35.0 (#176) | [@dependabot](https://github.com/dependabot)
- Bump github.com/aws/aws-sdk-go from 1.44.127 to 1.44.150 (#177) | [@dependabot](https://github.com/dependabot)
- KFH RDS Transform: refactor code into 'common' package (#180) | [@jharley](https://github.com/jharley)

## 3.0.0 2022-11-23

Introducing NEW [CloudFormation](https://github.com/honeycombio/cloudformation-integrations) & [Terraform](https://github.com/honeycombio/terraform-aws-integrations) support. 
Modules and templates are no longer maintained in this repo.

### Enhancement

-  Remove CloudFormation/Terraform & Rewrite README Significantly (#167) | [@dstrelau](https://github.com/dstrelau)

## 2.6.0 2022-11-21

### Features

- Cloudfront Lambda transforms (#166) | [@dstrelau](https://github.com/dstrelau)


## 2.5.0 2022-11-10

### Features

- RDS Lambda transforms (#160) | [@brookesargent](https://github.com/brookesargent)
- Embed service regexs into code (#150) | [@dstrelau](https://github.com/dstrelau)

### Fixed

- [s3] Protect against Intn panicing if sample rate is set to 0 (#148) | [@mjayaram](https://github.com/mjayaram)

### Maintenance

- publish every main build to S3 (#161) | [@dstrelau](https://github.com/dstrelau)
- Bump github.com/aws/aws-sdk-go from 1.44.109 to 1.44.127 (#158) | [@dependabot](https://github.com/dependabot)
- Bump github.com/stretchr/testify from 1.8.0 to 1.8.1 (#155) | [@dependabot](https://github.com/dependabot)
- Bump github.com/honeycombio/libhoney-go from 1.17.0 to 1.18.0 (#157) | [@dependabot](https://github.com/dependabot)
- maint: delete workflows for old board (#152) | [@vreynolds](https://github.com/vreynolds)
- maint: add release file (#149) | [@vreynolds](https://github.com/vreynolds)
- maint: add new project workflow (#147) | [@vreynolds](https://github.com/vreynolds)

## 2.4.5 2022-10-06

### Fixed

- [s3] fix bug in sampling & exception cases (#145) | [@lizthegrey](https://github.com/lizthegrey)

## 2.4.4 2022-10-06

### Fixed

- [s3] presample before parsing (#143) | [@lizthegrey](https://github.com/lizthegrey)

### Maintenance

- Bump github.com/aws/aws-sdk-go from 1.44.106 to 1.44.109 (#141)

## 2.4.3 2022-09-27

### Fixed

- Remove trailing "t" from default log_line_prefix (#136) | [@robbkidd](https://github.com/robbkidd)
- Update README to indicate audit logs are not supported (#132) | [@pkanal](https://github.com/pkanal)

### Maintenance

- Bump github.com/aws/aws-lambda-go from 1.27.1 to 1.34.1 (#128)
- Bump github.com/aws/aws-sdk-go from 1.43.36 to 1.44.106 (#138)
- Bump github.com/honeycombio/honeytail from 1.7.1 to 1.8.1 (#134)
- Bump github.com/honeycombio/libhoney-go from 1.15.8 to 1.17.0 (#137)
- Bump github.com/sirupsen/logrus from 1.8.1 to 1.9.0 (#127)
- Bump github.com/honeycombio/honeytail from 1.6.2 to 1.7.1 (#126)

## 2.4.2 2022-07-20

### Maintenance

- Re-release to fix OpenSSL CVE [@kentquirk](https://github.com/kentquirk)

## 2.4.1 2022-04-26

### Maintenance

- update ci image to cimg/go1.18 (#113) | [@JamieDanielson](https://github.com/JamieDanielson)
  - fixes openSSL CVE
- Bump github.com/stretchr/testify from 1.7.0 to 1.7.1 (#109) | [dependabot](https://github.com/dependabot)
- Bump github.com/aws/aws-sdk-go from 1.42.34 to 1.43.36 (#112) | [dependabot](https://github.com/dependabot)
- Bump github.com/honeycombio/honeytail from 1.6.0 to 1.6.1 (#102) | [dependabot](https://github.com/dependabot)

## 2.4.0 2022-02-09

### Features

- feat: rename fields for alb ingest (#104) | [@ryanking](https://github.com/ryanking)

## 2.3.1 2022-02-02

### Fixes

- fix: alb logs are not parsed (#97) | [@vreynolds](https://github.com/vreynolds)

### Maintenance

- gh: add re-triage workflow (#90) | [@vreynolds](https://github.com/vreynolds)
- docs: add a note about debugging (#82) | [@vreynolds](https://github.com/vreynolds)
- Bump github.com/aws/aws-sdk-go from 1.42.16 to 1.42.34 (#93)
- Bump github.com/aws/aws-lambda-go from 1.27.0 to 1.27.1 (#91)
- Bump github.com/aws/aws-sdk-go from 1.42.9 to 1.42.16 (#88)

## 2.3.0 2021-11-22

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
