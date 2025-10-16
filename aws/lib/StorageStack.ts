import * as cdk from 'aws-cdk-lib';
import * as s3 from 'aws-cdk-lib/aws-s3';
import { Construct } from 'constructs';

export class StorageStack extends cdk.Stack {
  public readonly contentBucket: s3.IBucket;

  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Create public S3 bucket for static content
    this.contentBucket = new s3.Bucket(this, 'SunsContentBucket', {
      bucketName: `suns-bz-content-${this.account}`,
      blockPublicAccess: new s3.BlockPublicAccess({
        blockPublicAcls: false,
        blockPublicPolicy: false,
        ignorePublicAcls: false,
        restrictPublicBuckets: false,
      }),
      publicReadAccess: true,
      websiteIndexDocument: 'index.html',
      websiteErrorDocument: '404.html',
      encryption: s3.BucketEncryption.S3_MANAGED,
      versioned: false,
      removalPolicy: cdk.RemovalPolicy.RETAIN,
      autoDeleteObjects: false,
    });

    new cdk.CfnOutput(this, 'BucketName', {
      value: this.contentBucket.bucketName,
      description: 'S3 Bucket Name',
      exportName: 'SunsContentBucketName',
    });

    new cdk.CfnOutput(this, 'BucketArn', {
      value: this.contentBucket.bucketArn,
      description: 'S3 Bucket ARN',
      exportName: 'SunsContentBucketArn',
    });

    new cdk.CfnOutput(this, 'BucketWebsiteUrl', {
      value: this.contentBucket.bucketWebsiteUrl,
      description: 'S3 Bucket Website URL',
    });
  }
}
