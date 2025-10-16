import * as cdk from 'aws-cdk-lib';
import * as cloudfront from 'aws-cdk-lib/aws-cloudfront';
import * as origins from 'aws-cdk-lib/aws-cloudfront-origins';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as acm from 'aws-cdk-lib/aws-certificatemanager';
import { Construct } from 'constructs';
import { config } from './config';

export interface EdgeStackProps extends cdk.StackProps {
  contentBucket: s3.IBucket;
  certificate: acm.ICertificate;
}

export class EdgeStack extends cdk.Stack {
  public readonly distribution: cloudfront.IDistribution;

  constructor(scope: Construct, id: string, props: EdgeStackProps) {
    super(scope, id, props);

    // Create CloudFront Distribution with public S3 bucket origin
    this.distribution = new cloudfront.Distribution(this, `${config.stackPrefix}Distribution`, {
      defaultBehavior: {
        origin: origins.S3BucketOrigin.withOriginAccessControl(props.contentBucket),
        viewerProtocolPolicy: cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
        allowedMethods: cloudfront.AllowedMethods.ALLOW_GET_HEAD_OPTIONS,
        cachedMethods: cloudfront.CachedMethods.CACHE_GET_HEAD_OPTIONS,
        compress: true,
        cachePolicy: cloudfront.CachePolicy.CACHING_OPTIMIZED,
      },
      domainNames: [config.domainName, `zq.${config.domainName}`],
      certificate: props.certificate,
      defaultRootObject: 'index.html',
      errorResponses: [
        {
          httpStatus: 404,
          responseHttpStatus: 404,
          responsePagePath: '/404.html',
          ttl: cdk.Duration.minutes(5),
        },
        {
          httpStatus: 403,
          responseHttpStatus: 403,
          responsePagePath: '/403.html',
          ttl: cdk.Duration.minutes(5),
        },
      ],
      minimumProtocolVersion: cloudfront.SecurityPolicyProtocol.TLS_V1_2_2021,
      httpVersion: cloudfront.HttpVersion.HTTP2_AND_3,
      priceClass: cloudfront.PriceClass.PRICE_CLASS_100,
      comment: `CloudFront distribution for ${config.domainName} and zq.${config.domainName}`,
    });

    new cdk.CfnOutput(this, 'DistributionId', {
      value: this.distribution.distributionId,
      description: 'CloudFront Distribution ID',
      exportName: `${config.stackPrefix}DistributionId`,
    });

    new cdk.CfnOutput(this, 'DistributionDomainName', {
      value: this.distribution.distributionDomainName,
      description: 'CloudFront Distribution Domain Name',
      exportName: `${config.stackPrefix}DistributionDomainName`,
    });
  }
}
