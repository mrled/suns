import * as cdk from 'aws-cdk-lib';
import * as cloudfront from 'aws-cdk-lib/aws-cloudfront';
import * as origins from 'aws-cdk-lib/aws-cloudfront-origins';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as acm from 'aws-cdk-lib/aws-certificatemanager';
import * as apigateway from 'aws-cdk-lib/aws-apigatewayv2';
import { Construct } from 'constructs';
import { config } from './config';

export interface EdgeStackProps extends cdk.StackProps {
  contentBucket: s3.IBucket;
  certificate: acm.ICertificate;
  webhookApi: apigateway.IHttpApi;
}

export class EdgeStack extends cdk.Stack {
  public readonly distribution: cloudfront.IDistribution;

  constructor(scope: Construct, id: string, props: EdgeStackProps) {
    super(scope, id, props);

    // Create CloudFront Distribution with S3 website endpoint
    // Using website endpoint instead of OAC to avoid cyclic dependency
    // (since bucket is already publicly accessible)
    const websiteOrigin = new origins.HttpOrigin(
      props.contentBucket.bucketWebsiteDomainName,
      {
        protocolPolicy: cloudfront.OriginProtocolPolicy.HTTP_ONLY,
      }
    );

    // Extract the domain from the API endpoint URL
    // API endpoint format: https://xxxxx.execute-api.region.amazonaws.com
    // We need to extract just the domain part
    const apiDomainName = cdk.Fn.select(2, cdk.Fn.split('/', props.webhookApi.apiEndpoint));

    const apiOrigin = new origins.HttpOrigin(apiDomainName, {
      protocolPolicy: cloudfront.OriginProtocolPolicy.HTTPS_ONLY,
      // No origin path needed since the API already has /v1/attest route
    });

    // Additional behaviors for API Gateway
    const additionalBehaviors: Record<string, cloudfront.BehaviorOptions> = {
      // Also catch any other /v1/* paths for future expansion
      '/v1/*': {
        origin: apiOrigin,
        viewerProtocolPolicy: cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
        allowedMethods: cloudfront.AllowedMethods.ALLOW_ALL,
        cachedMethods: cloudfront.CachedMethods.CACHE_GET_HEAD_OPTIONS,
        // Use caching disabled policy for API requests
        cachePolicy: cloudfront.CachePolicy.CACHING_DISABLED,
        // Use a policy that forwards headers except Host header for API Gateway
        // (ALL_VIEWER forwards Host header which causes API Gateway to return 403)
        originRequestPolicy: cloudfront.OriginRequestPolicy.ALL_VIEWER_EXCEPT_HOST_HEADER,
      },
    };

    this.distribution = new cloudfront.Distribution(this, `${config.stackPrefix}Distribution`, {
      defaultBehavior: {
        origin: websiteOrigin,
        viewerProtocolPolicy: cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
        allowedMethods: cloudfront.AllowedMethods.ALLOW_GET_HEAD_OPTIONS,
        cachedMethods: cloudfront.CachedMethods.CACHE_GET_HEAD_OPTIONS,
        compress: true,
        cachePolicy: cloudfront.CachePolicy.CACHING_OPTIMIZED,
      },
      additionalBehaviors,
      domainNames: [config.domainName, `zq.${config.domainName}`],
      certificate: props.certificate,
      defaultRootObject: 'index.html',
      // Commented out error responses to allow API errors to pass through
      // TODO: Consider implementing error responses that only apply to S3 content
      // errorResponses: [
      //   {
      //     httpStatus: 404,
      //     responseHttpStatus: 404,
      //     responsePagePath: '/404.html',
      //     ttl: cdk.Duration.minutes(5),
      //   },
      //   {
      //     httpStatus: 403,
      //     responseHttpStatus: 403,
      //     responsePagePath: '/403.html',
      //     ttl: cdk.Duration.minutes(5),
      //   },
      // ],
      minimumProtocolVersion: cloudfront.SecurityPolicyProtocol.TLS_V1_2_2021,
      httpVersion: cloudfront.HttpVersion.HTTP2_AND_3,
      priceClass: cloudfront.PriceClass.PRICE_CLASS_100,
      comment: `CloudFront distribution for ${config.domainName} and zq.${config.domainName} with API Gateway at /v1/attest and /v1/*`,
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

    new cdk.CfnOutput(this, 'WebhookEndpointViaCDN', {
      value: `https://${config.domainName}/v1/attest`,
      description: 'Webhook attest endpoint via CloudFront CDN',
    });
  }
}
