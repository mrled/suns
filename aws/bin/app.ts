#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { DnsZoneStack } from '../lib/dns-zone-stack';
import { CertStack } from '../lib/cert-stack';
import { StorageStack } from '../lib/storage-stack';
import { EdgeStack } from '../lib/edge-stack';
import { DnsStack } from '../lib/dns-stack';

const app = new cdk.App();

const account = process.env.CDK_DEFAULT_ACCOUNT;
const region = process.env.CDK_DEFAULT_REGION || 'us-east-1';

// DNS Zone Stack - can be in any region
const dnsZoneStack = new DnsZoneStack(app, 'SunsDnsZoneStack', {
  env: { account, region },
  description: 'Route53 hosted zone for suns.bz',
});

// Certificate Stack - MUST be in us-east-1 for CloudFront
const certStack = new CertStack(app, 'SunsCertStack', {
  env: { account, region: 'us-east-1' },
  description: 'ACM certificate for suns.bz (us-east-1 for CloudFront)',
  hostedZone: dnsZoneStack.hostedZone,
  crossRegionReferences: true,
});
certStack.addDependency(dnsZoneStack);

// Storage Stack - can be in any region
const storageStack = new StorageStack(app, 'SunsStorageStack', {
  env: { account, region },
  description: 'S3 bucket for suns.bz static content',
});

// Edge Stack - can be in any region (CloudFront is global)
const edgeStack = new EdgeStack(app, 'SunsEdgeStack', {
  env: { account, region },
  description: 'CloudFront distribution for suns.bz',
  contentBucket: storageStack.contentBucket,
  certificate: certStack.certificate,
  crossRegionReferences: true,
});
edgeStack.addDependency(storageStack);
edgeStack.addDependency(certStack);

// DNS Records Stack - must be in same region as hosted zone
const dnsStack = new DnsStack(app, 'SunsDnsStack', {
  env: { account, region },
  description: 'DNS records for suns.bz pointing to CloudFront',
  hostedZone: dnsZoneStack.hostedZone,
  distribution: edgeStack.distribution,
});
dnsStack.addDependency(dnsZoneStack);
dnsStack.addDependency(edgeStack);

app.synth();
