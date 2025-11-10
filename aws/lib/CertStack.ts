import * as cdk from "aws-cdk-lib";
import * as acm from "aws-cdk-lib/aws-certificatemanager";
import * as route53 from "aws-cdk-lib/aws-route53";
import { Construct } from "constructs";
import { config } from "./config";

export interface CertStackProps extends cdk.StackProps {
  hostedZone: route53.IHostedZone;
}

// Certificates MUST be in us-east-1 for CloudFront

export class CertStack extends cdk.Stack {
  public readonly certificate: acm.ICertificate;
  public readonly certificateV2: acm.ICertificate;

  constructor(scope: Construct, id: string, props: CertStackProps) {
    super(scope, id, props);

    // Original cert
    this.certificate = new acm.Certificate(
      this,
      `${config.stackPrefix}Certificate`,
      {
        domainName: config.domainName,
        subjectAlternativeNames: [`*.${config.domainName}`],
        validation: acm.CertificateValidation.fromDns(props.hostedZone),
      },
    );

    // New cert to add subjAlt
    this.certificateV2 = new acm.Certificate(
      this,
      `${config.stackPrefix}CertificateV2`,
      {
        domainName: config.domainName,
        subjectAlternativeNames: [
          `*.${config.domainName}`,
          `*.snus.${config.domainName}`,
        ],
        validation: acm.CertificateValidation.fromDns(props.hostedZone),
      },
    );

    new cdk.CfnOutput(this, "CertificateArn", {
      value: this.certificate.certificateArn,
      description: "ACM Certificate ARN",
    });
  }
}
