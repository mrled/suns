import * as cdk from "aws-cdk-lib";
import * as route53 from "aws-cdk-lib/aws-route53";
import { Construct } from "constructs";
import { config } from "./config";

export class DnsZoneStack extends cdk.Stack {
  public readonly hostedZone: route53.IHostedZone;

  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Create Route53 Hosted Zone
    this.hostedZone = new route53.PublicHostedZone(
      this,
      `${config.stackPrefix}HostedZone`,
      {
        zoneName: config.domainName,
        comment: `Hosted zone for ${config.domainName} domain`,
      },
    );

    // Output nameservers
    new cdk.CfnOutput(this, "HostedZoneId", {
      value: this.hostedZone.hostedZoneId,
      description: "Route53 Hosted Zone ID",
    });

    new cdk.CfnOutput(this, "NameServers", {
      value: cdk.Fn.join(", ", this.hostedZone.hostedZoneNameServers || []),
      description: "Name servers for the hosted zone",
    });
  }
}
