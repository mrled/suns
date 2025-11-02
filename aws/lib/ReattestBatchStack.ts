import * as cdk from "aws-cdk-lib";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as logs from "aws-cdk-lib/aws-logs";
import * as events from "aws-cdk-lib/aws-events";
import * as targets from "aws-cdk-lib/aws-events-targets";
import { Construct } from "constructs";
import { config, repositoryRoot } from "./config";
import * as path from "path";

export interface ReattestBatchStackProps extends cdk.StackProps {
  table: dynamodb.ITable;
  contentBucket: s3.IBucket;
}

export class ReattestBatchStack extends cdk.Stack {
  public readonly reattestBatchFunction: lambda.Function;

  constructor(scope: Construct, id: string, props: ReattestBatchStackProps) {
    super(scope, id, props);

    // Create Lambda function for batch re-attestation
    this.reattestBatchFunction = new lambda.Function(
      this,
      "ReattestBatchFunction",
      {
        runtime: lambda.Runtime.PROVIDED_AL2023,
        handler: "bootstrap",
        architecture: lambda.Architecture.ARM_64, // Graviton2
        code: lambda.Code.fromAsset(path.join(repositoryRoot, "symval"), {
          bundling: {
            image: lambda.Runtime.PROVIDED_AL2023.bundlingImage,
            command: [
              "sh",
              "-c",
              [
                "dnf install -y golang",
                "export GOOS=linux",
                "export GOARCH=arm64",
                "export CGO_ENABLED=0",
                "cd /asset-input",
                'go build -tags netgo -ldflags "-s -w -extldflags -static" -trimpath -o /asset-output/bootstrap ./cmd/reattestbatch',
              ].join(" && "),
            ],
            user: "root",
          },
        }),
        environment: {
          DYNAMODB_TABLE: props.table.tableName,
          S3_BUCKET: props.contentBucket.bucketName,
          S3_DATA_KEY: config.domainsDataKey || "records/domains.json",
        },
        timeout: cdk.Duration.minutes(5),
        memorySize: 256,
      },
    );

    // Grant DynamoDB permissions
    props.table.grantReadWriteData(this.reattestBatchFunction);

    // Grant S3 permissions
    props.contentBucket.grantReadWrite(this.reattestBatchFunction);

    // Set log retention on the auto-created log group
    new logs.LogRetention(this, "ReattestBatchFunctionLogRetention", {
      logGroupName: `/aws/lambda/${this.reattestBatchFunction.functionName}`,
      retention: logs.RetentionDays.ONE_WEEK,
    });

    // Create EventBridge rule to trigger the function on a schedule
    const rule = new events.Rule(this, "ReattestBatchScheduleRule", {
      schedule: events.Schedule.rate(cdk.Duration.hours(24)),
      description: "Trigger re-attestation batch process every day",
    });

    // Add the Lambda function as a target of the rule
    rule.addTarget(
      new targets.LambdaFunction(this.reattestBatchFunction, {
        retryAttempts: 2,
      }),
    );

    // Outputs
    new cdk.CfnOutput(this, "ReattestBatchFunctionArn", {
      value: this.reattestBatchFunction.functionArn,
      description: "Re-attestation Batch Lambda Function ARN",
      exportName: `${config.stackPrefix}ReattestBatchFunctionArn`,
    });

    new cdk.CfnOutput(this, "ReattestBatchFunctionName", {
      value: this.reattestBatchFunction.functionName,
      description: "Re-attestation Batch Lambda Function Name",
      exportName: `${config.stackPrefix}ReattestBatchFunctionName`,
    });

    new cdk.CfnOutput(this, "ScheduleRuleArn", {
      value: rule.ruleArn,
      description: "EventBridge Schedule Rule ARN",
      exportName: `${config.stackPrefix}ReattestBatchScheduleRuleArn`,
    });
  }
}
