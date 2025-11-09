import * as cdk from "aws-cdk-lib";
import * as cloudfront from "aws-cdk-lib/aws-cloudfront";
import * as origins from "aws-cdk-lib/aws-cloudfront-origins";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as acm from "aws-cdk-lib/aws-certificatemanager";
import * as apigateway from "aws-cdk-lib/aws-apigatewayv2";
import { Construct } from "constructs";
import { config } from "./config";

export interface EdgeStackProps extends cdk.StackProps {
  contentBucket: s3.IBucket;
  certificate: acm.ICertificate;
  httpApi: apigateway.IHttpApi;
}

export class EdgeStack extends cdk.Stack {
  public readonly distribution: cloudfront.IDistribution;

  constructor(scope: Construct, id: string, props: EdgeStackProps) {
    super(scope, id, props);

    // Create CloudFront Function for redirects
    const redirectFunctionCode = `
function handler(event) {
  var request = event.request;
  var headers = request.headers;
  var host = headers.host && headers.host.value;
  var uri = request.uri;

  // Rule 1: Redirect / to //:sdʇʇɥ (301 permanent redirect)
  if (uri === '/' || uri === '') {
    return {
      statusCode: 301,
      statusDescription: 'Moved Permanently',
      headers: {
        'location': { value: '//:sdʇʇɥ' },
        'cache-control': { value: 'max-age=3600' }
      }
    };
  }

  // Rule 2: Redirect /:sdʇʇɥ to //:sdʇʇɥ (301 permanent redirect)
  if (uri === '/:sdʇʇɥ') {
    return {
      statusCode: 301,
      statusDescription: 'Moved Permanently',
      headers: {
        'location': { value: '//:sdʇʇɥ' },
        'cache-control': { value: 'max-age=3600' }
      }
    };
  }

  // Rule 3: Redirect //:sdʇʇɥ/ (with trailing slash) to //:sdʇʇɥ (without trailing slash)
  if (uri === '//:sdʇʇɥ/') {
    return {
      statusCode: 301,
      statusDescription: 'Moved Permanently',
      headers: {
        'location': { value: '//:sdʇʇɥ' },
        'cache-control': { value: 'max-age=3600' }
      }
    };
  }

  // Check if the host is suns.bz (without zq subdomain)
  if (host === 'suns.bz') {
    // Construct the new URL with zq subdomain, preserving path and query string
    var newUrl = 'https://zq.suns.bz' + uri;

    // Add query string if present
    if (request.querystring && Object.keys(request.querystring).length > 0) {
      var queryParts = [];
      for (var key in request.querystring) {
        var param = request.querystring[key];
        if (param.multiValue) {
          for (var i = 0; i < param.multiValue.length; i++) {
            queryParts.push(encodeURIComponent(key) + '=' + encodeURIComponent(param.multiValue[i].value));
          }
        } else {
          queryParts.push(encodeURIComponent(key) + '=' + encodeURIComponent(param.value));
        }
      }
      newUrl += '?' + queryParts.join('&');
    }

    // Return a 301 permanent redirect
    return {
      statusCode: 301,
      statusDescription: 'Moved Permanently',
      headers: {
        'location': { value: newUrl },
        'cache-control': { value: 'max-age=3600' }
      }
    };
  }

  // For zq.suns.bz or any other host, continue with the request
  return request;
}
    `.trim();

    const redirectFunction = new cloudfront.Function(
      this,
      "RedirectToZqFunction",
      {
        code: cloudfront.FunctionCode.fromInline(redirectFunctionCode),
        comment:
          "Handles redirects: root to //:sdʇʇɥ, /:sdʇʇɥ to //:sdʇʇɥ, removes trailing slash, and redirects suns.bz to zq.suns.bz",
      },
    );

    // Create CloudFront Distribution with S3 website endpoint
    // Using website endpoint instead of OAC to avoid cyclic dependency
    // (since bucket is already publicly accessible)
    // This origin points to the /www prefix in S3 for Hugo site content
    const websiteOrigin = new origins.HttpOrigin(
      props.contentBucket.bucketWebsiteDomainName,
      {
        protocolPolicy: cloudfront.OriginProtocolPolicy.HTTP_ONLY,
        originPath: "/www", // Serve Hugo content from /www prefix in S3
      },
    );

    // Create an origin for the records data (no /www prefix, serves from bucket root)
    const recordsOrigin = new origins.HttpOrigin(
      props.contentBucket.bucketWebsiteDomainName,
      {
        protocolPolicy: cloudfront.OriginProtocolPolicy.HTTP_ONLY,
        // No origin path - serve directly from bucket root to access /records/*
      },
    );

    // Extract the domain from the API endpoint URL
    // API endpoint format: https://xxxxx.execute-api.region.amazonaws.com
    // We need to extract just the domain part
    const apiDomainName = cdk.Fn.select(
      2,
      cdk.Fn.split("/", props.httpApi.apiEndpoint),
    );

    const apiOrigin = new origins.HttpOrigin(apiDomainName, {
      protocolPolicy: cloudfront.OriginProtocolPolicy.HTTPS_ONLY,
      // No origin path needed since the API Gateway handles routing internally
    });

    // Additional behaviors for API Gateway and records
    const additionalBehaviors: Record<string, cloudfront.BehaviorOptions> = {
      // Route all /api/* paths to the API Gateway Lambda
      "/api/*": {
        origin: apiOrigin,
        viewerProtocolPolicy: cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
        allowedMethods: cloudfront.AllowedMethods.ALLOW_ALL,
        cachedMethods: cloudfront.CachedMethods.CACHE_GET_HEAD_OPTIONS,
        // Use caching disabled policy for API requests
        cachePolicy: cloudfront.CachePolicy.CACHING_DISABLED,
        // Use a policy that forwards headers except Host header for API Gateway
        // (ALL_VIEWER forwards Host header which causes API Gateway to return 403)
        originRequestPolicy:
          cloudfront.OriginRequestPolicy.ALL_VIEWER_EXCEPT_HOST_HEADER,
        functionAssociations: [
          {
            function: redirectFunction,
            eventType: cloudfront.FunctionEventType.VIEWER_REQUEST,
          },
        ],
      },
      // Route /records/* paths to the S3 bucket root (for domains.json and future records)
      "/records/*": {
        origin: recordsOrigin,
        viewerProtocolPolicy: cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
        allowedMethods: cloudfront.AllowedMethods.ALLOW_GET_HEAD_OPTIONS,
        cachedMethods: cloudfront.CachedMethods.CACHE_GET_HEAD_OPTIONS,
        compress: true,
        // Cache for a short duration since this data updates periodically
        cachePolicy: new cloudfront.CachePolicy(this, "RecordsCachePolicy", {
          cachePolicyName: `${config.stackPrefix}RecordsCachePolicy`,
          defaultTtl: cdk.Duration.minutes(5),
          maxTtl: cdk.Duration.hours(1),
          minTtl: cdk.Duration.seconds(0),
          enableAcceptEncodingGzip: true,
          enableAcceptEncodingBrotli: true,
        }),
        // Add CORS headers to allow cross-origin requests
        responseHeadersPolicy: new cloudfront.ResponseHeadersPolicy(
          this,
          "RecordsResponseHeadersPolicy",
          {
            responseHeadersPolicyName: `${config.stackPrefix}RecordsResponseHeadersPolicy`,
            corsBehavior: {
              accessControlAllowOrigins: ["*"],
              accessControlAllowHeaders: ["*"],
              accessControlAllowMethods: ["GET", "HEAD", "OPTIONS"],
              accessControlAllowCredentials: false,
              originOverride: true,
            },
          },
        ),
        functionAssociations: [
          {
            function: redirectFunction,
            eventType: cloudfront.FunctionEventType.VIEWER_REQUEST,
          },
        ],
      },
    };

    this.distribution = new cloudfront.Distribution(
      this,
      `${config.stackPrefix}Distribution`,
      {
        defaultBehavior: {
          origin: websiteOrigin,
          viewerProtocolPolicy:
            cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
          allowedMethods: cloudfront.AllowedMethods.ALLOW_GET_HEAD_OPTIONS,
          cachedMethods: cloudfront.CachedMethods.CACHE_GET_HEAD_OPTIONS,
          compress: true,
          cachePolicy: cloudfront.CachePolicy.CACHING_OPTIMIZED,
          functionAssociations: [
            {
              function: redirectFunction,
              eventType: cloudfront.FunctionEventType.VIEWER_REQUEST,
            },
          ],
        },
        additionalBehaviors,
        domainNames: [config.domainName, `zq.${config.domainName}`],
        certificate: props.certificate,
        defaultRootObject: "index.html",
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
        comment: `CloudFront distribution for ${config.domainName} and zq.${config.domainName} with API Gateway at /api/* and records at /records/*`,
      },
    );

    new cdk.CfnOutput(this, "DistributionId", {
      value: this.distribution.distributionId,
      description: "CloudFront Distribution ID",
    });

    new cdk.CfnOutput(this, "DistributionDomainName", {
      value: this.distribution.distributionDomainName,
      description: "CloudFront Distribution Domain Name",
    });

    new cdk.CfnOutput(this, "WebhookEndpointViaCDN", {
      value: `https://${config.domainName}/api/v1/attest`,
      description: "Webhook attest endpoint via CloudFront CDN",
    });

    new cdk.CfnOutput(this, "DomainsDataEndpoint", {
      value: `https://${config.domainName}/${config.domainsDataKey}`,
      description: "Domains data JSON endpoint via CloudFront CDN",
    });
  }
}
