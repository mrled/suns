import * as cdk from "aws-cdk-lib";
import * as acm from "aws-cdk-lib/aws-certificatemanager";
import * as route53 from "aws-cdk-lib/aws-route53";
import { Construct } from "constructs";
import { config } from "./config";

export interface CertStackProps extends cdk.StackProps {
  hostedZone: route53.IHostedZone;
}

export class CertStack extends cdk.Stack {
  public readonly certificate: acm.ICertificate;

  constructor(scope: Construct, id: string, props: CertStackProps) {
    super(scope, id, props);

    // Create ACM Certificate for domain and wildcard
    // Must be in us-east-1 for CloudFront
    this.certificate = new acm.Certificate(
      this,
      `${config.stackPrefix}Certificate`,
      {
        domainName: config.domainName,
        subjectAlternativeNames: [`*.${config.domainName}`],
        validation: acm.CertificateValidation.fromDns(props.hostedZone),
      },
    );

    new cdk.CfnOutput(this, "CertificateArn", {
      value: this.certificate.certificateArn,
      description: "ACM Certificate ARN",
    });
  }
}
