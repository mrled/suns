#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { config } from '../lib/config';
import { DnsZoneStack } from '../lib/DnsZoneStack';
import { CertStack } from '../lib/CertStack';
import { StorageStack } from '../lib/StorageStack';
import { EdgeStack } from '../lib/EdgeStack';
import { DnsStack } from '../lib/DnsStack';
import { DynamoDbStack } from '../lib/DynamoDbStack';
import { WebhookStack } from '../lib/WebhookStack';

const app = new cdk.App();

const account = process.env.CDK_DEFAULT_ACCOUNT;
const region = config.deployRegion;

// DNS Zone Stack - can be in any region
const dnsZoneStack = new DnsZoneStack(app, `${config.stackPrefix}DnsZoneStack`, {
  env: { account, region },
  description: `Route53 hosted zone for ${config.domainName}`,
});

// Certificate Stack - MUST be in us-east-1 for CloudFront
const certStack = new CertStack(app, `${config.stackPrefix}CertStack`, {
  env: { account, region: config.acmRegion },
  description: `ACM certificate for ${config.domainName} (${config.acmRegion} for CloudFront)`,
  hostedZone: dnsZoneStack.hostedZone,
  crossRegionReferences: true,
});
certStack.addDependency(dnsZoneStack);

// Storage Stack - can be in any region
const storageStack = new StorageStack(app, `${config.stackPrefix}StorageStack`, {
  env: { account, region },
  description: `S3 bucket for ${config.domainName} static content`,
});

// DynamoDB Stack - can be in any region
const dynamoDbStack = new DynamoDbStack(app, `${config.stackPrefix}DynamoDbStack`, {
  env: { account, region },
  description: `DynamoDB table for ${config.domainName}`,
});

// Webhook Stack - Lambda + API Gateway for attestation
const webhookStack = new WebhookStack(app, `${config.stackPrefix}WebhookStack`, {
  env: { account, region },
  description: `Webhook Lambda and API Gateway for ${config.domainName}`,
  table: dynamoDbStack.table,
});
webhookStack.addDependency(dynamoDbStack);

// Edge Stack - can be in any region (CloudFront is global)
const edgeStack = new EdgeStack(app, `${config.stackPrefix}EdgeStack`, {
  env: { account, region },
  description: `CloudFront distribution for ${config.domainName}`,
  contentBucket: storageStack.contentBucket,
  certificate: certStack.certificate,
  webhookApi: webhookStack.api,
  crossRegionReferences: true,
});
edgeStack.addDependency(storageStack);
edgeStack.addDependency(certStack);
edgeStack.addDependency(webhookStack);

// DNS Records Stack - must be in same region as hosted zone
const dnsStack = new DnsStack(app, `${config.stackPrefix}DnsStack`, {
  env: { account, region },
  description: `DNS records for ${config.domainName} pointing to CloudFront`,
  hostedZone: dnsZoneStack.hostedZone,
  distribution: edgeStack.distribution,
});
dnsStack.addDependency(dnsZoneStack);
dnsStack.addDependency(edgeStack);

app.synth();
