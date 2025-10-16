import * as cdk from 'aws-cdk-lib';
import * as route53 from 'aws-cdk-lib/aws-route53';
import { Construct } from 'constructs';

export class DnsZoneStack extends cdk.Stack {
  public readonly hostedZone: route53.IHostedZone;

  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Create Route53 Hosted Zone for suns.bz
    this.hostedZone = new route53.PublicHostedZone(this, 'SunsHostedZone', {
      zoneName: 'suns.bz',
      comment: 'Hosted zone for suns.bz domain',
    });

    // Output nameservers
    new cdk.CfnOutput(this, 'HostedZoneId', {
      value: this.hostedZone.hostedZoneId,
      description: 'Route53 Hosted Zone ID',
      exportName: 'SunsHostedZoneId',
    });

    new cdk.CfnOutput(this, 'NameServers', {
      value: cdk.Fn.join(', ', this.hostedZone.hostedZoneNameServers || []),
      description: 'Name servers for the hosted zone',
    });
  }
}
