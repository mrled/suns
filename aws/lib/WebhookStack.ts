import * as cdk from 'aws-cdk-lib';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as apigateway from 'aws-cdk-lib/aws-apigatewayv2';
import * as apigatewayIntegrations from 'aws-cdk-lib/aws-apigatewayv2-integrations';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import { Construct } from 'constructs';
import { config, repositoryRoot } from './config';
import * as path from 'path';

export interface WebhookStackProps extends cdk.StackProps {
  table: dynamodb.ITable;
}

export class WebhookStack extends cdk.Stack {
  public readonly api: apigateway.HttpApi;
  public readonly webhookFunction: lambda.Function;

  constructor(scope: Construct, id: string, props: WebhookStackProps) {
    super(scope, id, props);

    // Create Lambda function for webhook
    this.webhookFunction = new lambda.Function(this, 'WebhookFunction', {
      runtime: lambda.Runtime.PROVIDED_AL2023,
      handler: 'bootstrap',
      architecture: lambda.Architecture.ARM_64, // Graviton2
      code: lambda.Code.fromAsset(path.join(repositoryRoot, 'symval'), {
        bundling: {
          image: lambda.Runtime.PROVIDED_AL2023.bundlingImage,
          command: [
            'sh',
            '-c',
            [
              'dnf install -y golang',
              'export GOOS=linux',
              'export GOARCH=arm64',
              'export CGO_ENABLED=0',
              'cd /asset-input',
              'go build -tags netgo -ldflags "-s -w -extldflags -static" -trimpath -o /asset-output/bootstrap ./cmd/webhook',
            ].join(' && '),
          ],
          user: 'root',
        },
      }),
      environment: {
        // AWS_REGION: this.region, // This is set by the Lambda runtime and cannot be overridden
        DYNAMODB_TABLE: props.table.tableName,
      },
      timeout: cdk.Duration.seconds(5),
      memorySize: 128,
    });

    // Grant DynamoDB permissions
    props.table.grantReadWriteData(this.webhookFunction);

    // Create HTTP API Gateway
    this.api = new apigateway.HttpApi(this, 'WebhookApi', {
      apiName: `${config.stackPrefix}WebhookApi`,
      description: 'HTTP API for SUNS webhook attestation',
      corsPreflight: {
        allowOrigins: ['*'],
        allowMethods: [
          apigateway.CorsHttpMethod.GET,
          apigateway.CorsHttpMethod.POST,
          apigateway.CorsHttpMethod.PUT,
          apigateway.CorsHttpMethod.DELETE,
          apigateway.CorsHttpMethod.PATCH,
        ],
        allowHeaders: ['Content-Type', 'Authorization'],
      },
    });

    // Create Lambda integration
    const lambdaIntegration = new apigatewayIntegrations.HttpLambdaIntegration(
      'WebhookIntegration',
      this.webhookFunction
    );

    // Add route for all /api/* paths - this will proxy all requests to Lambda
    // The Lambda function will handle routing internally
    this.api.addRoutes({
      path: '/api/{proxy+}',
      methods: [
        apigateway.HttpMethod.GET,
        apigateway.HttpMethod.POST,
        apigateway.HttpMethod.PUT,
        apigateway.HttpMethod.DELETE,
        apigateway.HttpMethod.PATCH,
      ],
      integration: lambdaIntegration,
    });

    // Outputs
    new cdk.CfnOutput(this, 'ApiUrl', {
      value: this.api.apiEndpoint,
      description: 'Webhook API Gateway URL',
      exportName: `${config.stackPrefix}WebhookApiUrl`,
    });

    new cdk.CfnOutput(this, 'ApiId', {
      value: this.api.apiId,
      description: 'Webhook API Gateway ID',
      exportName: `${config.stackPrefix}WebhookApiId`,
    });

    new cdk.CfnOutput(this, 'FunctionArn', {
      value: this.webhookFunction.functionArn,
      description: 'Webhook Lambda Function ARN',
      exportName: `${config.stackPrefix}WebhookFunctionArn`,
    });

    new cdk.CfnOutput(this, 'AttestEndpoint', {
      value: `${this.api.apiEndpoint}/api/v1/attest`,
      description: 'Direct API Gateway webhook attest endpoint URL',
    });
  }
}
