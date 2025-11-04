# SUNS Infrastructure

AWS CDK infrastructure for suns.bz. 5 stacks: DNS zone, ACM cert (must be us-east-1), S3 storage, CloudFront, and DNS records.

## Setup & Deploy

```sh
# Install dependencies
npm install

# Bootstrap CDK (first time only)
# This must be done in all regions we use.
# IMPORTANT: ACM certificate MUST be in us-east-1 (CloudFront requirement)
npx cdk bootstrap aws://ACCOUNT-ID/us-east-1

# Optional: bootstrap other regions if needed
npx cdk bootstrap aws://ACCOUNT-ID/REGION

# Review stacks
npm run synth

# Deploy all stacks (handles dependencies automatically)
npm run deploy

# Or deploy individual stacks
npx cdk deploy SunsDnsZoneStack # ... etc

# After DNS zone deploys, update nameservers at your domain registrar
# Outputs will show: SunsDnsZoneStack.NameServers = ns-xxxx.awsdns...

# Other useful commands
npx cdk diff                    # Compare deployed vs current state
npx cdk list                    # List all stacks
npx cdk destroy STACK_NAME      # Remove specific stack
npm run build                   # Compile TypeScript
npm run watch                   # Watch mode
```

Deploying the website is done with `hugo deploy`,
which copies files to S3 and also invalidates the CloudFront content cache.

## Stack Architecture

1. **SunsDnsZoneStack** - Route53 hosted zone for suns.bz
2. **SunsCertStack** - ACM certificate for suns.bz and \*.suns.bz (us-east-1 only)
3. **SunsStorageStack** - S3 bucket for static content and DynamoDB table
4. **SunsEdgeStack** - CloudFront distribution serving from S3
5. **SunsDnsStack** - A/AAAA records pointing to CloudFront
6. **SunsHttpApiStack** - HTTP API Gateway with Lambda function
   - Lambda: `SunsHttpApiFunction` - Handles API requests at /api/\* paths
7. **SunsStreamerStack** - DynamoDB Streams processor
   - Lambda: `SunsStreamerFunction` - Processes DynamoDB stream events
8. **SunsReattestBatchStack** - Scheduled batch processing
   - Lambda: `SunsReattestBatchFunction` - Runs daily re-attestation batch job
9. **SunsMonitoringStack** - CloudWatch dashboards and alarms

Dependencies: DNS Zone → Cert + DNS Records → Edge → Storage → HttpApi/Streamer/ReattestBatch → Monitoring

## Notes

- Certificate stack must be in us-east-1 (CloudFront requirement)
- CloudFront deployment takes 15-30 minutes
- Certificate validation may take a few minutes after nameserver update
- S3 bucket is public for CloudFront access
- HTTPS enforced, TLS 1.2+ minimum
- Route53 costs ~$0.50/month, CloudFront/S3 pay-as-you-go, ACM free

## CloudWatch Logs

Lambda functions have predictable names for easy log access:

```sh
# HTTP API function logs (handles /api/* requests)
aws logs tail /aws/lambda/SunsHttpApiFunction --follow

# DynamoDB Streams processor logs
aws logs tail /aws/lambda/SunsStreamerFunction --follow

# Batch re-attestation job logs (runs daily)
aws logs tail /aws/lambda/SunsReattestBatchFunction --follow

# View recent logs without following
aws logs tail /aws/lambda/SunsHttpApiFunction --since 1h

# Filter logs by pattern
aws logs filter-log-events \
  --log-group-name /aws/lambda/SunsHttpApiFunction \
  --filter-pattern "ERROR"
```
