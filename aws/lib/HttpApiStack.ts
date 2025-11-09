import * as cdk from "aws-cdk-lib";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as apigateway from "aws-cdk-lib/aws-apigatewayv2";
import * as apigatewayIntegrations from "aws-cdk-lib/aws-apigatewayv2-integrations";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as logs from "aws-cdk-lib/aws-logs";
import { Construct } from "constructs";
import { config } from "./config";
import { getUnifiedLambdaCode } from "./LambdaCode";

export interface HttpApiStackProps extends cdk.StackProps {
  table: dynamodb.ITable;
}

export class HttpApiStack extends cdk.Stack {
  public readonly api: apigateway.HttpApi;
  public readonly apiFunction: lambda.Function;

  constructor(scope: Construct, id: string, props: HttpApiStackProps) {
    super(scope, id, props);

    // Create Lambda function for HTTP API
    this.apiFunction = new lambda.Function(this, "HttpApiFunction", {
      runtime: lambda.Runtime.PROVIDED_AL2023,
      handler: "bootstrap",
      architecture: lambda.Architecture.ARM_64, // Graviton2
      functionName: config.functionNames.httpApi,
      code: getUnifiedLambdaCode(this),
      environment: {
        // AWS_REGION: this.region, // This is set by the Lambda runtime and cannot be overridden
        LAMBDA_HANDLER: "httpapi",
        DYNAMODB_TABLE: props.table.tableName,
      },
      timeout: cdk.Duration.seconds(5),
      memorySize: 128,
    });

    // Grant DynamoDB permissions
    props.table.grantReadWriteData(this.apiFunction);

    // Set log retention on the auto-created log group
    new logs.LogRetention(this, "HttpApiFunctionLogRetention", {
      logGroupName: `/aws/lambda/${this.apiFunction.functionName}`,
      retention: logs.RetentionDays.ONE_WEEK,
    });

    // Create HTTP API Gateway
    this.api = new apigateway.HttpApi(this, "HttpApi", {
      apiName: `${config.stackPrefix}HttpApi`,
      description: "HTTP API for SUNS attestation",
      corsPreflight: {
        allowOrigins: ["*"],
        allowMethods: [
          apigateway.CorsHttpMethod.GET,
          apigateway.CorsHttpMethod.POST,
          apigateway.CorsHttpMethod.PUT,
          apigateway.CorsHttpMethod.DELETE,
          apigateway.CorsHttpMethod.PATCH,
        ],
        allowHeaders: ["Content-Type", "Authorization"],
      },
    });

    // Create Lambda integration
    const lambdaIntegration = new apigatewayIntegrations.HttpLambdaIntegration(
      "HttpApiIntegration",
      this.apiFunction,
    );

    // Add route for all /api/* paths - this will proxy all requests to Lambda
    // The Lambda function will handle routing internally
    this.api.addRoutes({
      path: "/api/{proxy+}",
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
    new cdk.CfnOutput(this, "ApiUrl", {
      value: this.api.apiEndpoint,
      description: "HTTP API Gateway URL",
    });

    new cdk.CfnOutput(this, "ApiId", {
      value: this.api.apiId,
      description: "HTTP API Gateway ID",
    });

    new cdk.CfnOutput(this, "AttestEndpoint", {
      value: `${this.api.apiEndpoint}/api/v1/attest`,
      description: "Direct API Gateway attest endpoint URL",
    });
  }
}
