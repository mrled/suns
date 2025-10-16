import * as cdk from 'aws-cdk-lib';
import * as route53 from 'aws-cdk-lib/aws-route53';
import * as targets from 'aws-cdk-lib/aws-route53-targets';
import * as cloudfront from 'aws-cdk-lib/aws-cloudfront';
import { Construct } from 'constructs';

export interface DnsStackProps extends cdk.StackProps {
  hostedZone: route53.IHostedZone;
  distribution: cloudfront.IDistribution;
}

export class DnsStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: DnsStackProps) {
    super(scope, id, props);

    // Create Route53 A record for suns.bz pointing to CloudFront
    new route53.ARecord(this, 'SunsARecord', {
      zone: props.hostedZone,
      recordName: 'suns.bz',
      target: route53.RecordTarget.fromAlias(
        new targets.CloudFrontTarget(props.distribution)
      ),
    });

    // Create Route53 AAAA record for suns.bz (IPv6)
    new route53.AaaaRecord(this, 'SunsAaaaRecord', {
      zone: props.hostedZone,
      recordName: 'suns.bz',
      target: route53.RecordTarget.fromAlias(
        new targets.CloudFrontTarget(props.distribution)
      ),
    });

    // Create Route53 A record for zq.suns.bz pointing to CloudFront
    new route53.ARecord(this, 'ZqARecord', {
      zone: props.hostedZone,
      recordName: 'zq',
      target: route53.RecordTarget.fromAlias(
        new targets.CloudFrontTarget(props.distribution)
      ),
    });

    // Create Route53 AAAA record for zq.suns.bz (IPv6)
    new route53.AaaaRecord(this, 'ZqAaaaRecord', {
      zone: props.hostedZone,
      recordName: 'zq',
      target: route53.RecordTarget.fromAlias(
        new targets.CloudFrontTarget(props.distribution)
      ),
    });
  }
}
