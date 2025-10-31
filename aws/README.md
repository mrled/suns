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

# Upload content to S3 (get bucket name from stack outputs)
aws s3 sync ./content s3://BUCKET-NAME/

# Invalidate CloudFront cache after content changes
# `hugo deploy` actually will do this if invalidateCDN is set to true (the defautl), so no need to run this ourselves:
aws cloudfront create-invalidation --distribution-id DISTRIBUTION-ID --paths "/*"

# Other useful commands
npx cdk diff                    # Compare deployed vs current state
npx cdk list                    # List all stacks
npx cdk destroy STACK_NAME      # Remove specific stack
npm run build                   # Compile TypeScript
npm run watch                   # Watch mode
```

## Stack Architecture

1. SunsDnsZoneStack - Route53 hosted zone for suns.bz
2. SunsCertStack - ACM certificate for suns.bz and *.suns.bz (us-east-1 only)
3. SunsStorageStack - S3 bucket for static content
4. SunsEdgeStack - CloudFront distribution serving from S3
5. SunsDnsStack - A/AAAA records pointing to CloudFront

Dependencies: DNS Zone → Cert + DNS Records → Edge → Storage

## Notes

- Certificate stack must be in us-east-1 (CloudFront requirement)
- CloudFront deployment takes 15-30 minutes
- Certificate validation may take a few minutes after nameserver update
- S3 bucket is public for CloudFront access
- HTTPS enforced, TLS 1.2+ minimum
- Route53 costs ~$0.50/month, CloudFront/S3 pay-as-you-go, ACM free

## CloudWatch logs

```sh
# Read logs from the function that backs the /api URL path
aws logs tail /aws/lambda/SunsApiFunction --region us-east-2
```
