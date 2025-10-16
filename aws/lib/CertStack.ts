import * as cdk from 'aws-cdk-lib';
import * as acm from 'aws-cdk-lib/aws-certificatemanager';
import * as route53 from 'aws-cdk-lib/aws-route53';
import { Construct } from 'constructs';

export interface CertStackProps extends cdk.StackProps {
  hostedZone: route53.IHostedZone;
}

export class CertStack extends cdk.Stack {
  public readonly certificate: acm.ICertificate;

  constructor(scope: Construct, id: string, props: CertStackProps) {
    super(scope, id, props);

    // Create ACM Certificate for suns.bz and *.suns.bz
    // Must be in us-east-1 for CloudFront
    this.certificate = new acm.Certificate(this, 'SunsCertificate', {
      domainName: 'suns.bz',
      subjectAlternativeNames: ['*.suns.bz'],
      validation: acm.CertificateValidation.fromDns(props.hostedZone),
    });

    new cdk.CfnOutput(this, 'CertificateArn', {
      value: this.certificate.certificateArn,
      description: 'ACM Certificate ARN',
      exportName: 'SunsCertificateArn',
    });
  }
}
