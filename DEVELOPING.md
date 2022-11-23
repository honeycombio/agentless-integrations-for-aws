# Developing the integrations

## Testing local changes

0. Have a relevant AWS input resource ready (e.g. a CloudWatch log group with some test logs)
1. Create an S3 bucket that will hold the integrations source code and save the bucket name (e.g. `agentless-test`)
2. Build the integrations: `./build.sh`
   1. If you're on an ARM-based development machine like an Apple M-series, use `GOARCH=amd64 ./build.sh`, since the `go1.x` Lambda runtime only supports x86 binaries.
3. Sync the integrations code into the S3 bucket: `aws s3 cp pkg/ingest-handlers.zip s3://agentless-test/ingest-handlers.zip`
4. Manually configure the lambda, or use either the Terraform modules or CloudFormation stacks, specifying your the S3 bucket and path you used above.
