import * as cdk from "aws-cdk-lib";
import * as route53 from "aws-cdk-lib/aws-route53";
import * as targets from "aws-cdk-lib/aws-route53-targets";
import * as cloudfront from "aws-cdk-lib/aws-cloudfront";
import { Construct } from "constructs";

export interface DnsStackProps extends cdk.StackProps {
  hostedZone: route53.IHostedZone;
  distribution: cloudfront.IDistribution;
}

export class DnsStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: DnsStackProps) {
    super(scope, id, props);

    // Create Route53 A record for suns.bz pointing to CloudFront
    new route53.ARecord(this, "SunsARecord", {
      zone: props.hostedZone,
      recordName: "suns.bz",
      target: route53.RecordTarget.fromAlias(
        new targets.CloudFrontTarget(props.distribution),
      ),
    });

    // Create Route53 AAAA record for suns.bz (IPv6)
    new route53.AaaaRecord(this, "SunsAaaaRecord", {
      zone: props.hostedZone,
      recordName: "suns.bz",
      target: route53.RecordTarget.fromAlias(
        new targets.CloudFrontTarget(props.distribution),
      ),
    });

    // Create Route53 A record for zq.suns.bz pointing to CloudFront
    new route53.ARecord(this, "ZqARecord", {
      zone: props.hostedZone,
      recordName: "zq",
      target: route53.RecordTarget.fromAlias(
        new targets.CloudFrontTarget(props.distribution),
      ),
    });

    // Create Route53 AAAA record for zq.suns.bz (IPv6)
    new route53.AaaaRecord(this, "ZqAaaaRecord", {
      zone: props.hostedZone,
      recordName: "zq",
      target: route53.RecordTarget.fromAlias(
        new targets.CloudFrontTarget(props.distribution),
      ),
    });

    // Create Route53 A record for zb.snus.suns.bz pointing to CloudFront
    new route53.ARecord(this, "ZbSnusARecord", {
      zone: props.hostedZone,
      recordName: "zb.snus",
      target: route53.RecordTarget.fromAlias(
        new targets.CloudFrontTarget(props.distribution),
      ),
    });

    // Create Route53 AAAA record for zb.snus.suns.bz (IPv6)
    new route53.AaaaRecord(this, "ZbSnusAaaaRecord", {
      zone: props.hostedZone,
      recordName: "zb.snus",
      target: route53.RecordTarget.fromAlias(
        new targets.CloudFrontTarget(props.distribution),
      ),
    });

    // Create TXT record for _suns.zq.suns.bz
    new route53.TxtRecord(this, "SunsZqTxtRecord", {
      zone: props.hostedZone,
      recordName: "_suns.zq",
      values: [
        // flip180
        "v1:b:Adg2VY1R7OIZHMz/Z1uk1lOCe2DtwzXYTpaBHpDEtkU=:AjlAANu+Rl8TYqK0CvGy+S6NIswJmRuEsYSn43MG+2s=",
      ],
      ttl: cdk.Duration.minutes(1),
    });

    // Attestation for _suns.zb.snus.suns.bz
    new route53.TxtRecord(this, "SunsZbSnusTxtRecord", {
      zone: props.hostedZone,
      recordName: "_suns.zb.snus",
      values: [
        // palindrome
        "v1:a:1TcHFCGUO8ipNQfOLGZ5xs2nnhEsKOqIQkYyXkE9vcE=:jgvt4BDuPrTOIado+FqXJh2eMMF9invGPxf0gdonycM=",
      ],
      ttl: cdk.Duration.minutes(1),
    });
  }
}
