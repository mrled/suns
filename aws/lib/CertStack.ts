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

    // Certificate
    // This can never be changed after creation;
    // if we try to change it, a new cert will be created, which creates a new ARN,
    // which will prevent CloudFormation from updating any stacks that depend on this one.
    // Instead, create a new cert with a different name,
    // update dependent stacks to use the new cert,
    // then delete this cert once they're all updated.
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
  }
}
