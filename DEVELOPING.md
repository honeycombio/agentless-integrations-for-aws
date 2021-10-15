# Developing the integrations

## Testing local changes

0. Have a relevant AWS input resource ready (e.g. a CloudWatch log group with some test logs)
1. Create an S3 bucket that will hold the integrations source code and save the bucket name (e.g. `agentless-test`)
2. Build the integrations: `./build.sh`
3. Sync the integrations code into the S3 bucket: `aws s3 cp pkg/ingest-handlers.zip s3://agentless-test/ingest-handlers.zip`
4. Update the relevant template under `./templates` to use the test S3 bucket, for example update `cloudwatch-logs-json.yml` to have these properties under the resource handler:
   1. S3Bucket: agentless-test
   2. S3Key: ingest-handlers.zip
5. In AWS Console, navigate to CloudFormation --> Create Stack
6. Pick `Upload a template file` and upload the updated template from above (e.g. `./templates/cloudwatch-logs-json.yml`)
7. Follow the rest of configuration instructions from public docs