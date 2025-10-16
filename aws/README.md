# SUNS Infrastructure

AWS CDK infrastructure for suns.bz domain, split into multiple stacks for modularity and independent deployment.

## Architecture

The infrastructure is organized into 5 separate CDK stacks:

### 1. DNS Zone Stack (`SunsDnsZoneStack`)
- **Route53 Hosted Zone**: `suns.bz`
- Contains only the hosted zone definition
- Can be deployed in any region

### 2. Certificate Stack (`SunsCertStack`)
- **ACM Certificate**: SSL/TLS certificate for `suns.bz` and `*.suns.bz`
- **Region**: MUST be in `us-east-1` (required by CloudFront)
- Uses DNS validation via the hosted zone

### 3. Storage Stack (`SunsStorageStack`)
- **S3 Bucket**: Public bucket for static content with website hosting enabled
- Can be deployed in any region
- Bucket name: `suns-bz-content-{account-id}`

### 4. Edge Stack (`SunsEdgeStack`)
- **CloudFront Distribution**: Global CDN serving content from S3
- **Domains**:
  - `suns.bz` (apex domain)
  - `zq.suns.bz` (subdomain)
- Can be deployed in any region (CloudFront is global)

### 5. DNS Records Stack (`SunsDnsStack`)
- **Route53 A/AAAA Records**: Points domains to CloudFront
- Must be in same region as hosted zone

## Stack Dependencies

```
SunsDnsZoneStack
├─> SunsCertStack (requires hosted zone)
├─> SunsDnsStack (requires hosted zone)
    └─> SunsEdgeStack (requires distribution)
        ├─> SunsStorageStack (requires bucket)
        └─> SunsCertStack (requires certificate)
```

## Prerequisites

1. AWS CLI configured with appropriate credentials
2. Node.js (v18 or later)
3. Domain registered (suns.bz) - you'll need to update nameservers after deployment

## Installation

```bash
cd aws
npm install
```

## Deployment

### 1. Bootstrap CDK (first time only)

Bootstrap in us-east-1 (required for certificate stack):

```bash
npx cdk bootstrap aws://ACCOUNT-ID/us-east-1
```

If deploying to a different region, bootstrap that region too:

```bash
npx cdk bootstrap aws://ACCOUNT-ID/REGION
```

Or use your default account/region:

```bash
npx cdk bootstrap
```

### 2. Review all stacks

```bash
npm run synth
```

### 3. Deploy all stacks

Deploy all stacks at once:

```bash
npm run deploy
```

Or deploy individual stacks:

```bash
npx cdk deploy SunsDnsZoneStack
npx cdk deploy SunsCertStack
npx cdk deploy SunsStorageStack
npx cdk deploy SunsEdgeStack
npx cdk deploy SunsDnsStack
```

Or deploy multiple specific stacks:

```bash
npx cdk deploy SunsDnsZoneStack SunsCertStack SunsStorageStack
```

### 4. Update Domain Nameservers

After deploying the DNS Zone Stack, you'll see output with nameservers:

```
Outputs:
SunsDnsZoneStack.NameServers = ns-xxxx.awsdns-xx.com, ns-xxxx.awsdns-xx.net, ...
```

Update your domain registrar to use these nameservers for suns.bz.

## Post-Deployment

### Upload content to S3

Get the bucket name from outputs:

```bash
aws s3 sync ./your-content-folder s3://BUCKET-NAME/
```

Or use the AWS Console to upload files.

### Invalidate CloudFront cache

After uploading new content:

```bash
aws cloudfront create-invalidation --distribution-id DISTRIBUTION-ID --paths "/*"
```

Replace `DISTRIBUTION-ID` with the value from stack outputs.

## Stack Outputs

### SunsDnsZoneStack
- `HostedZoneId`: Route53 Hosted Zone ID
- `NameServers`: Nameservers to configure at your domain registrar

### SunsCertStack
- `CertificateArn`: ACM Certificate ARN

### SunsStorageStack
- `BucketName`: S3 bucket name for content
- `BucketArn`: S3 bucket ARN
- `BucketWebsiteUrl`: S3 bucket website URL

### SunsEdgeStack
- `DistributionId`: CloudFront distribution ID
- `DistributionDomainName`: CloudFront distribution domain name

### SunsDnsStack
- (Creates DNS records, no outputs)

## Benefits of Multi-Stack Architecture

1. **Independent Deployment**: Deploy or update individual components without affecting others
2. **Region Flexibility**: Deploy resources in optimal regions (except cert in us-east-1)
3. **Clear Separation**: Each stack has a single, well-defined responsibility
4. **Easier Debugging**: Issues are isolated to specific stacks
5. **Selective Destruction**: Can destroy stacks independently (be careful with dependencies)

## Configuration

- **Certificate Stack**: Always deployed to `us-east-1` (CloudFront requirement)
- **Other Stacks**: Default to `us-east-1` or `CDK_DEFAULT_REGION`
- **Cross-Region References**: Enabled for cert and edge stacks

### Security Features

- S3 bucket is public for direct CloudFront access
- HTTPS redirect for all HTTP requests
- TLS 1.2+ minimum protocol version
- Content encryption at rest with S3 managed keys

## Useful Commands

- `npm run build` - Compile TypeScript to JavaScript
- `npm run watch` - Watch for changes and compile
- `npm run synth` - Synthesize CloudFormation template for all stacks
- `npm run deploy` - Deploy all stacks to AWS
- `npx cdk deploy STACK_NAME` - Deploy specific stack
- `npx cdk diff` - Compare deployed stacks with current state
- `npx cdk destroy STACK_NAME` - Remove specific stack
- `npx cdk destroy --all` - Remove all stacks (be careful!)
- `npx cdk list` - List all stacks

## Cost Considerations

- Route53 Hosted Zone: ~$0.50/month
- CloudFront: Pay per request and data transfer
- S3: Pay per storage and requests
- ACM Certificate: Free

## Troubleshooting

### Certificate validation pending

The ACM certificate uses DNS validation. After deployment, it may take a few minutes for the certificate to be validated and issued once the nameservers are updated.

### CloudFront takes time to deploy

CloudFront distributions can take 15-30 minutes to fully deploy globally.

### Content not updating

Remember to invalidate the CloudFront cache after uploading new content to S3.

### Cross-region reference errors

If you encounter cross-region reference errors, ensure you've bootstrapped both us-east-1 and your deployment region.

### Stack dependencies

Always deploy stacks in order or use `npx cdk deploy --all` to handle dependencies automatically.
