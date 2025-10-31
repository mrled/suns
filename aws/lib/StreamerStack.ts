import * as cdk from 'aws-cdk-lib';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as s3 from 'aws-cdk-lib/aws-s3';
import * as lambdaEventSources from 'aws-cdk-lib/aws-lambda-event-sources';
import { Construct } from 'constructs';
import { config, repositoryRoot } from './config';
import * as path from 'path';

export interface StreamerStackProps extends cdk.StackProps {
  table: dynamodb.ITable;
  contentBucket: s3.IBucket; // Required: the content bucket to write domains data to
}

export class StreamerStack extends cdk.Stack {
  public readonly streamerFunction: lambda.Function;

  constructor(scope: Construct, id: string, props: StreamerStackProps) {
    super(scope, id, props);

    // Create Lambda function for DynamoDB Streams processing
    this.streamerFunction = new lambda.Function(this, 'StreamerFunction', {
      runtime: lambda.Runtime.PROVIDED_AL2023,
      handler: 'bootstrap',
      architecture: lambda.Architecture.ARM_64, // Graviton2
      functionName: `${config.stackPrefix}StreamerFunction`,
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
              'go build -tags netgo -ldflags "-s -w -extldflags -static" -trimpath -o /asset-output/bootstrap ./cmd/streamer',
            ].join(' && '),
          ],
          user: 'root',
        },
      }),
      environment: {
        DYNAMODB_TABLE: props.table.tableName,
        S3_BUCKET: props.contentBucket.bucketName,
        S3_DATA_KEY: config.domainsDataKey, // Pass the key path as an env var
      },
      timeout: cdk.Duration.seconds(30), // Longer timeout for stream processing
      memorySize: 256, // More memory for stream processing
      reservedConcurrentExecutions: 1, // IMPORTANT: Only allow one instance to run at once
    });

    // Grant DynamoDB permissions
    props.table.grantReadWriteData(this.streamerFunction);
    props.table.grantStreamRead(this.streamerFunction);

    // Grant S3 permissions to write to the content bucket
    props.contentBucket.grantWrite(this.streamerFunction);

    // Add DynamoDB Streams event source
    // We need to cast the table to Table type to access streams
    const table = props.table as dynamodb.Table;

    // Add the DynamoDB Streams event source to the Lambda function
    this.streamerFunction.addEventSource(new lambdaEventSources.DynamoEventSource(table, {
      startingPosition: lambda.StartingPosition.LATEST,
      batchSize: 10, // Process up to 10 records at once
      bisectBatchOnError: true, // Split the batch on error to isolate bad records
      retryAttempts: 3, // Retry failed batches up to 3 times
      reportBatchItemFailures: true, // Report individual item failures
      parallelizationFactor: 1, // Process one shard at a time (since we have concurrency=1)
    }));

    // Outputs
    new cdk.CfnOutput(this, 'StreamerFunctionArn', {
      value: this.streamerFunction.functionArn,
      description: 'DynamoDB Streams Lambda Function ARN',
      exportName: `${config.stackPrefix}StreamerFunctionArn`,
    });

    new cdk.CfnOutput(this, 'StreamerFunctionName', {
      value: this.streamerFunction.functionName,
      description: 'DynamoDB Streams Lambda Function Name',
      exportName: `${config.stackPrefix}StreamerFunctionName`,
    });

    new cdk.CfnOutput(this, 'DomainsDataUrl', {
      value: `https://${config.domainName}/${config.domainsDataKey}`,
      description: 'URL to fetch domains data JSON file',
    });
  }
}